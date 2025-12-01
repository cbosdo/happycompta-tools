// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// Employee describes the data of an employee.
type Employee struct {
	ID        string
	Lastname  string
	Firstname string
	Active    bool
}

// GetID is needed for Employee to implement the Party interface.
func (e *Employee) GetID() string {
	return e.ID
}

// IsValid indicates if the required data are available.
func (e *Employee) IsValid() bool {
	return e.ID != "" && e.Firstname != "" && e.Lastname != ""
}

// ListEmployees returns a list of all employees.
func (c *Client) ListEmployees() (employees []Employee, err error) {
	values := url.Values{}
	values.Set("statut_salarie", "-1")
	values.Set("site_id", "0")
	values.Set("sexe", "")
	values.Set("situation_familiale", "0")
	req, err := http.NewRequest("POST", url_base+"/salaries/ajax_table", strings.NewReader(values.Encode()))
	if err != nil {
		err = fmt.Errorf("failed to create the request: %s", err)
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := c.client.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to get the list of employees: %s", err)
		return
	}

	defer resp.Body.Close()
	return parseEmployeesResponse(resp.Body)
}

func parseEmployeesResponse(r io.Reader) (employees []Employee, err error) {
	var content struct {
		View string `json:"view"`
	}

	jsonDecoder := json.NewDecoder(r)
	if err = jsonDecoder.Decode(&content); err != nil {
		err = fmt.Errorf("failed to decode JSON: %s", err)
		return
	}

	htmlContent := content.View
	if htmlContent == "" {
		return
	}

	htmlReader := strings.NewReader(htmlContent)
	doc, err := html.ParseWithOptions(htmlReader, html.ParseOptionEnableScripting(false))
	if err != nil {
		err = fmt.Errorf("failed to parse the html employees table: %s", err)
		return
	}

	return parseEmployeesTable(doc)
}

func parseEmployeesTable(doc *html.Node) (employees []Employee, err error) {
	const (
		columnActive    = 2
		columnLastname  = 6
		columnFirstname = 7
		columnsActions  = 11
	)

	var currentEmployee *Employee
	var isInsideTbody bool
	var tdCount int

	// Function to traverse the DOM
	var traverseTree func(*html.Node)
	traverseTree = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "tbody" {
				isInsideTbody = true
			} else if isInsideTbody && n.Data == "tr" {
				// Start of a new employee row
				currentEmployee = &Employee{}
				tdCount = 0
			} else if isInsideTbody && n.Data == "td" {
				tdCount++

				if tdCount == columnActive {
					currentEmployee.Active = findClassText(n, "hide") == "1"
				}

				if tdCount == columnLastname {
					currentEmployee.Lastname = html.UnescapeString(extractTextContent(n))
				} else if tdCount == columnFirstname {
					currentEmployee.Firstname = html.UnescapeString(extractTextContent(n))
				}

				if tdCount == columnsActions {
					currentEmployee.ID = parseEmployeeID(n)
				}
			}

			for c := n.FirstChild; c != nil; c = c.NextSibling {
				traverseTree(c)
			}

			if isInsideTbody && n.Data == "tr" && currentEmployee != nil {
				if currentEmployee.IsValid() {
					employees = append(employees, *currentEmployee)
				}
				currentEmployee = nil
			} else if n.Data == "tbody" {
				isInsideTbody = false
			}
		} else {
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				traverseTree(c)
			}
		}

	}

	traverseTree(doc)
	return
}

// Regex to extract the ID from the 'edit' URL in the action column
var employeeIDRegex = regexp.MustCompile(`\/salaries\/edit\/(\d+)`)

// parseEmployeeID extracts the ID from the 'edit' URL in the last column.
//
// e.g., "https://app.happy-compta.fr/salaries/edit/123456" -> "123456"
func parseEmployeeID(node *html.Node) string {
	var traverseLink func(*html.Node) string
	traverseLink = func(t *html.Node) string {
		if t.Type == html.ElementNode && t.Data == "a" {
			for _, a := range t.Attr {
				if a.Key == "href" {
					match := employeeIDRegex.FindStringSubmatch(a.Val)
					if len(match) > 1 {
						return match[1]
					}
				}
			}
		}
		for c := t.FirstChild; c != nil; c = c.NextSibling {
			if id := traverseLink(c); id != "" {
				return id
			}
		}
		return ""
	}
	return traverseLink(node)
}
