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
		defer writer.Close()
		defer formWriter.Close()

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
			defer file.Close()

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
		io.Copy(io.Discard, reader)
		return fmt.Errorf("HTTP POST failed: %w", err)
	}
	defer resp.Body.Close()

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

	defer resp.Body.Close()

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
