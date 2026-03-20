// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// Date format constant (DD/MM/YYYY is a common format in the happy-compta)
const DateLayout = "02/01/2006"

// AllocationLine represents one line in the allocation of an entry.
type AllocationLine struct {
	CategoryID int
	Amount     float64
	Stock      int
}

// Party is an interface for an entry target.
type Party interface {
	// GetID returns the identifier of the party.
	GetID() string
}

// Entry represents an entry in the bookkeeping system.
type Entry struct {
	ID            string
	Period        string
	Kind          Kind
	Date          time.Time
	Name          string
	Budget        Budget
	Allocation    []AllocationLine
	Party         Party
	PaymentMethod PaymentMethod
	Account       Account
	Comment       string
	Receipts      []string
}

// ListEntries returns all the entries for a given period.
func (c *Client) ListEntries(periodID string) (result []Entry, err error) {
	// TODO Allow more filtering
	values := url.Values{}
	values.Set("statut", "toutes_operations")
	values.Set("type", "type")
	values.Set("budget", "0")
	values.Set("compte_id", "0")
	values.Set("method_paiement", "0")
	values.Set("cheque", "")
	values.Set("category_id", "0")
	values.Set("exercice_id", periodID)
	values.Set("begin", "")
	values.Set("end", "")
	values.Set("montant", "")
	values.Set("fournisseur_id", "0")
	values.Set("personne_id", "0")
	values.Set("pieces_jointes", "avec_sans_pj")
	req, err := http.NewRequest("POST", url_base+"/ajax/list_operations", strings.NewReader(values.Encode()))
	if err != nil {
		err = fmt.Errorf("failed to create the request: %s", err)
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := c.client.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to get the list of entries: %s", err)
		return
	}

	defer func() { _ = resp.Body.Close() }()
	doc, err := parseHtmlViewResponse(resp.Body)
	if err != nil {
		return
	}
	urls := getEntriesURLs(doc)
	for _, url := range urls {
		// TODO Implements virements
		if strings.Contains(url, "virements-internes") {
			continue
		}
		var entry Entry
		entry, err = c.getEntry(url)
		if err != nil {
			return
		}
		result = append(result, entry)
	}
	return
}

func (c *Client) getEntry(url string) (entry Entry, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		err = fmt.Errorf("failed to create the request: %s", err)
		return
	}
	resp, err := c.client.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to get the entry details: %s", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	return parseEntryResponse(resp.Body)
}

// parseEntryResponse parses the operation data.
// Mind that the employee / provider only have the ID set.
func parseEntryResponse(r io.Reader) (entry Entry, err error) {
	doc, err := html.ParseWithOptions(r, html.ParseOptionEnableScripting(false))
	if err != nil {
		return
	}

	// 1. Extract the JSON string from the <script> tag
	opData, err := extractOperationJSON(doc)
	if err != nil {
		return entry, err
	}

	// 2. Map JSON fields to the Entry struct
	entry.Name = opData.Name
	entry.Comment = opData.RemarquesLibres
	entry.Period = fmt.Sprintf("%d", opData.ExerciceID)
	entry.Kind = NewKind(opData.Type)
	entry.Budget = NewBudget(opData.Budget)
	entry.PaymentMethod = PaymentMethod(opData.MethodPaiement)
	entry.Account = Account{ID: opData.CompteID}

	// Parse Date
	if opData.Date != "" {
		entry.Date, _ = time.Parse("2006-01-02", opData.Date)
	}

	// 3. Handle Party (Provider or Employee) using IDs from JSON
	if opData.FournisseurID != nil {
		entry.Party = &Provider{ID: fmt.Sprintf("%v", opData.FournisseurID)}
	} else if opData.PersonneID != 0 {
		entry.Party = &Employee{ID: fmt.Sprintf("%d", opData.PersonneID)}
	}

	// 4. Map Allocations (Ventilations)
	for _, v := range opData.Ventilations {
		entry.Allocation = append(entry.Allocation, AllocationLine{
			CategoryID: v.CategoryID,
			Amount:     v.Amount,
			Stock:      v.Stock,
		})
	}

	// 5. Handle Multiple Receipts from filename_temp
	if opData.FilenameTemp != "" {
		// Files are semi-colon separated in this field
		entry.Receipts = strings.Split(opData.FilenameTemp, ";")
	}

	entry.ID = fmt.Sprintf("%s%06d", opData.IdentifiantPC, opData.NumeroPC)

	return entry, nil
}

// Internal struct matching the JSON structure in the HTML script
type jsonOperation struct {
	Name            string `json:"name"`
	Date            string `json:"date"`
	Type            string `json:"type"`
	Budget          int    `json:"budget"`
	ExerciceID      int    `json:"exercice_id"`
	CompteID        int    `json:"compte_id"`
	MethodPaiement  int    `json:"method_paiement"`
	FournisseurID   any    `json:"fournisseur_id"` // Can be null
	PersonneID      int    `json:"personne_id"`
	RemarquesLibres string `json:"remarques_libres"`
	FilenameTemp    string `json:"filename_temp"`
	Ventilations    []struct {
		CategoryID int     `json:"category_id"`
		Amount     float64 `json:"amount"`
		Stock      int     `json:"stock"`
	} `json:"ventilations"`
	IdentifiantPC string `json:"identifiant_pc"`
	NumeroPC      int    `json:"numero_pc"`
}

func extractOperationJSON(n *html.Node) (*jsonOperation, error) {
	for c := range n.Descendants() {
		if c.Type == html.ElementNode && c.Data == "script" {
			content := extractTextContent(c)
			if strings.Contains(content, "const operation = JSON.parse(String(\"") {
				// Extract the content between String(" and "))
				start := strings.Index(content, "String(\"") + 8
				end := strings.Index(content, "\"));\n")
				if start > 7 && end > start {
					jsonStr := content[start:end]
					// Unescape the string as it is stored as a JS string literal
					jsonStr = strings.ReplaceAll(jsonStr, `\"`, `"`)
					jsonStr = strings.ReplaceAll(jsonStr, `\\/`, `/`)
					jsonStr = strings.ReplaceAll(jsonStr, `\\`, `\`)

					var op jsonOperation
					if err := json.Unmarshal([]byte(jsonStr), &op); err != nil {
						return nil, fmt.Errorf("failed to unmarshal script JSON: %w", err)
					}
					return &op, nil
				}
			}
		}

	}
	return nil, fmt.Errorf("operation script not found")
}

// getEntriesURLs traverses the HTML tree and returns all hrefs
// containing the string "/operations/edit/"
func getEntriesURLs(n *html.Node) []string {
	var results []string
	var crawler func(*html.Node)

	crawler = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" && strings.Contains(a.Val, "/operations/edit/") {
					results = append(results, a.Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			crawler(c)
		}
	}
	crawler(n)
	return results
}

// AddEntry adds a new entry to the bookkeeping system.
func (c *Client) AddEntry(operation *Entry) error {
	entryID, entryIDNumber, err := c.getNextEntryNumber(operation.Budget, operation.Kind)
	if err != nil {
		return err
	}

	token, err := c.getToken(url_base + "/operations/create/depenses")
	if err != nil {
		return err
	}

	reader, writer := io.Pipe()
	formWriter := multipart.NewWriter(writer)

	go func() {
		defer func() { _ = writer.Close() }()
		defer func() { _ = formWriter.Close() }()

		if err := formWriter.WriteField("_token", token); err != nil {
			writer.CloseWithError(fmt.Errorf("error writing _token: %w", err))
			return
		}
		if err := formWriter.WriteField("exercice_id", operation.Period); err != nil {
			writer.CloseWithError(fmt.Errorf("error writing exercice_id: %w", err))
			return
		}

		if err := formWriter.WriteField("type", operation.Kind.String()); err != nil {
			writer.CloseWithError(fmt.Errorf("error writing type: %w", err))
			return
		}
		if err := formWriter.WriteField("budget", strconv.Itoa(int(operation.Budget))); err != nil {
			writer.CloseWithError(fmt.Errorf("error writing budget: %w", err))
			return
		}
		if err := formWriter.WriteField("date", operation.Date.Format(DateLayout)); err != nil {
			writer.CloseWithError(fmt.Errorf("error writing date: %w", err))
			return
		}
		if err := formWriter.WriteField("name", operation.Name); err != nil {
			writer.CloseWithError(fmt.Errorf("error writing name: %w", err))
			return
		}

		for _, line := range operation.Allocation {
			if err := formWriter.WriteField("category_id[]", strconv.Itoa(line.CategoryID)); err != nil {
				writer.CloseWithError(fmt.Errorf("error writing category_id[]: %w", err))
				return
			}
			amountStr := fmt.Sprintf("%.2f", line.Amount)
			amount := bytes.Replace([]byte(amountStr), []byte("."), []byte(","), 1)
			if err := formWriter.WriteField("amount[]", string(amount)); err != nil {
				writer.CloseWithError(fmt.Errorf("error writing amount[]: %w", err))
				return
			}
			if line.Stock != 0 {
				if err := formWriter.WriteField("stock[]", strconv.Itoa(line.Stock)); err != nil {
					writer.CloseWithError(fmt.Errorf("error writing stock[]: %w", err))
					return
				}
			} else {
				// Write an empty stock if none set
				if err := formWriter.WriteField("stock[]", ""); err != nil {
					writer.CloseWithError(fmt.Errorf("error writing empty stock[]: %w", err))
					return
				}
			}

			// TODO Handle the preorder date feature
			if err := formWriter.WriteField("date_remise_precommande", ""); err != nil {
				writer.CloseWithError(fmt.Errorf("error writing date_remise_precommande: %w", err))
				return
			}
			// This is field is set, but what is it used for?
			if err := formWriter.WriteField("ventilation_id[]", ""); err != nil {
				writer.CloseWithError(fmt.Errorf("error writing ventilation_id[]: %w", err))
				return
			}
		}

		providerID := "0"
		employeeID := "0"

		if _, ok := operation.Party.(*Provider); ok {
			if err := formWriter.WriteField("activateFournisseur", "on"); err != nil {
				writer.CloseWithError(fmt.Errorf("error writing activateSalarie: %w", err))
				return
			}
			providerID = operation.Party.GetID()
		} else if _, ok := operation.Party.(*Employee); ok {
			if err := formWriter.WriteField("activateSalarie", "on"); err != nil {
				writer.CloseWithError(fmt.Errorf("error writing activateSalarie: %w", err))
				return
			}
			employeeID = operation.Party.GetID()
		}

		if err := formWriter.WriteField("fournisseur_id", providerID); err != nil {
			writer.CloseWithError(fmt.Errorf("error writing default fournisseur_id: %w", err))
			return
		}
		if err := formWriter.WriteField("personne_id", employeeID); err != nil {
			writer.CloseWithError(fmt.Errorf("error writing default personne_id: %w", err))
			return
		}

		if err := formWriter.WriteField("method_paiement", strconv.Itoa(int(operation.PaymentMethod))); err != nil {
			writer.CloseWithError(fmt.Errorf("error writing method_paiement: %w", err))
			return
		}
		if err := formWriter.WriteField("compte_id", strconv.Itoa(operation.Account.ID)); err != nil {
			writer.CloseWithError(fmt.Errorf("error writing compte_id: %w", err))
			return
		}

		// File attachments (Receipts)
		for _, filePath := range operation.Receipts {
			file, err := os.Open(filePath)
			if err != nil {
				writer.CloseWithError(fmt.Errorf("error opening file %s: %w", filePath, err))
				return
			}
			defer func() { _ = file.Close() }()

			filename := filepath.Base(filePath)

			part, err := formWriter.CreateFormFile("fichiers[]", filename)
			if err != nil {
				writer.CloseWithError(fmt.Errorf("error creating form file part for %s: %w", filename, err))
				return
			}

			if _, err := io.Copy(part, file); err != nil {
				writer.CloseWithError(fmt.Errorf("error writing file content for %s: %w", filename, err))
				return
			}
		}

		if err := formWriter.WriteField("identifiant_pc", entryID); err != nil {
			writer.CloseWithError(fmt.Errorf("error writing identifiant_pc: %w", err))
			return
		}
		if err := formWriter.WriteField("numero_pc", entryIDNumber); err != nil {
			writer.CloseWithError(fmt.Errorf("error writing numero_pc: %w", err))
			return
		}

		// TODO Features not supported yet
		if err := formWriter.WriteField("nom_invite", ""); err != nil {
			writer.CloseWithError(fmt.Errorf("error writing nom_invite: %w", err))
			return
		}
		if err := formWriter.WriteField("prenom_invite", ""); err != nil {
			writer.CloseWithError(fmt.Errorf("error writing prenom_invite: %w", err))
			return
		}

		if err := formWriter.WriteField("no_cheque", ""); err != nil {
			writer.CloseWithError(fmt.Errorf("error writing no_cheque: %w", err))
			return
		}
		if err := formWriter.WriteField("banque", ""); err != nil {
			writer.CloseWithError(fmt.Errorf("error writing banque: %w", err))
			return
		}
		if err := formWriter.WriteField("date_remise_souhaitee", ""); err != nil {
			writer.CloseWithError(fmt.Errorf("error writing date_remise_souhaitee: %w", err))
			return
		}

		// Activation switches, may be they can be dropped
		if err := formWriter.WriteField("activateUpload", "on"); err != nil {
			writer.CloseWithError(fmt.Errorf("error writing activateUpload: %w", err))
			return
		}
		if err := formWriter.WriteField("activateRemarques", "on"); err != nil {
			writer.CloseWithError(fmt.Errorf("error writing activateRemarques: %w", err))
			return
		}

		// Static fields
		if err := formWriter.WriteField("confirm", "0"); err != nil {
			writer.CloseWithError(fmt.Errorf("error writing confirm: %w", err))
			return
		}
		if err := formWriter.WriteField("submit_value", "enregistrer"); err != nil {
			writer.CloseWithError(fmt.Errorf("error writing confirm: %w", err))
			return
		}

		if err := formWriter.Close(); err != nil {
			writer.CloseWithError(fmt.Errorf("error closing form writer: %w", err))
		}
	}()

	c.followRedirects(false)
	resp, err := c.client.Post(url_base+"/operations/store", formWriter.FormDataContentType(), reader)
	c.followRedirects(true)
	if err != nil {
		_, _ = io.Copy(io.Discard, reader)
		return fmt.Errorf("HTTP POST failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusFound {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(responseBody))
	}

	return nil
}

func (c *Client) getNextEntryNumber(budget Budget, kind Kind) (id string, number string, err error) {
	values := url.Values{}
	values.Set("operationId", "0")
	values.Set("operationType", kind.String())
	values.Set("budget", fmt.Sprintf("%d", int(budget)))
	req, err := http.NewRequest("POST", url_base+"/ajax/get-numero-pc", strings.NewReader(values.Encode()))
	if err != nil {
		err = fmt.Errorf("failed to create the request: %s", err)
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := c.client.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to get the next entry ID: %s", err)
		return
	}

	defer func() { _ = resp.Body.Close() }()

	type resultType struct {
		ID     string `json:"identifiant"`
		Number string `json:"numero"`
	}
	var result resultType

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("failed to read the response: %s", err)
		return
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		err = fmt.Errorf("failed to parse JSON response: %s", err)
		return
	}
	id = result.ID
	number = result.Number
	return
}
