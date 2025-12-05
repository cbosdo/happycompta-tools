// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"fmt"
	"io"
	"net/http"

	"golang.org/x/net/html"
)

type Provider struct {
	ID       string
	Name     string
	Address  string
	ZipCode  string
	City     string
	Phone    string
	Email    string
	Comment  string
	Archived bool
}

// GetID is needed for Provider to implement the Party interface.
func (p *Provider) GetID() string {
	return p.ID
}

// ListProviders queries the data of all the providers of the organization, included archived ones.
func (c *Client) ListProviders() (providers []Provider, err error) {
	resp, err := c.client.Get(url_base + "/fournisseurs/index/archiv%C3%A9s")
	if err != nil {
		err = fmt.Errorf("failed to get the providers: %s", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to get the providers, got %d status code", resp.StatusCode)
		return
	}

	return parseProviders(resp.Body)
}

func parseProviders(r io.Reader) (providers []Provider, err error) {
	doc, err := html.Parse(r)
	if err != nil {
		err = fmt.Errorf("failed to parse HTML: %w", err)
		return
	}

	tbody := findNodeWithTagName(doc, "tbody")

	if tbody == nil {
		err = fmt.Errorf("could not find the table listing the providers")
		return
	}

	rowIndex := 0

	const (
		columnName    = 0
		columnAddress = 1
		columnZipCode = 2
		columnCity    = 3
		columnPhone   = 4
		columnEmail   = 5
		columnComment = 6
		columnActions = 8
	)

	// Iterate through <tr> nodes in <tbody>
	for row := tbody.FirstChild; row != nil; row = row.NextSibling {
		if row.Type != html.ElementNode || row.Data != "tr" {
			continue
		}
		rowIndex++

		cells := []*html.Node{}
		for cell := row.FirstChild; cell != nil; cell = cell.NextSibling {
			if cell.Type == html.ElementNode && cell.Data == "td" {
				cells = append(cells, cell)
			}
		}

		if len(cells) < 9 {
			continue
		}

		var provider Provider

		provider.ID = extractIDFromActionsCell(cells[columnActions])
		provider.Name = extractTextContent(cells[columnName])
		provider.Address = extractTextContent(cells[columnAddress])
		provider.ZipCode = extractTextContent(cells[columnZipCode])
		provider.City = extractTextContent(cells[columnCity])
		provider.Phone = extractTextContent(cells[columnPhone])
		provider.Email = extractTextContent(cells[columnEmail])
		provider.Comment = extractTextContent(cells[columnComment])

		unarchiveBtn := findNodeWithKeyValueAttr(cells[columnActions], "data-archive", "1")
		provider.Archived = unarchiveBtn != nil

		providers = append(providers, provider)
	}
	return
}
