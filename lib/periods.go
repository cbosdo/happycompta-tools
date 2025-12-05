// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"golang.org/x/net/html"
)

type Period struct {
	ID     string
	Status PeriodStatus
	Start  time.Time
	End    time.Time
}

// ListPeriods gets the data of all the accounting periods of the organization.
func (c *Client) ListPeriods() (periods []Period, err error) {
	resp, err := c.client.Get(url_base + "/exercices/index")
	if err != nil {
		err = fmt.Errorf("failed to get the periods: %s", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to get the periods, got %d status code", resp.StatusCode)
		return
	}

	return parsePeriods(resp.Body)
}

// extractIDFromActionsCell searches the actions for tag with the data-id attribute and returns that value.
func extractIDFromActionsCell(cell *html.Node) string {
	targetNode := findNodeWithAttr(cell, "data-id")

	if targetNode != nil {
		return getAttr(targetNode, "data-id")
	}
	return ""
}

// extractStatusFromStatusCell traverses the status cell to find a hidden span.
func extractStatusFromStatusCell(cell *html.Node) (status PeriodStatus, err error) {
	reStatus := regexp.MustCompile(`" \. (\d) \. "`)

	hiddenSpan := findNodeWithKeyValueAttr(cell, "class", "hidden")

	if hiddenSpan != nil {
		for c := hiddenSpan.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.TextNode {
				statusMatch := reStatus.FindStringSubmatch(c.Data)
				if len(statusMatch) < 2 {
					err = fmt.Errorf("could not extract status number from text node: %s", c.Data)
					return
				}

				statusStr := statusMatch[1]
				var statusInt int
				statusInt, err = strconv.Atoi(statusStr)
				if err != nil {
					err = fmt.Errorf("failed to convert status '%s' to integer: %w", statusStr, err)
					return
				}
				status = NewPeriodStatus(statusInt)
				return
			}
		}
	}

	return 0, fmt.Errorf("could not find the hidden status span structure")
}

// parsePeriods reads the periods from HTML content.
func parsePeriods(r io.Reader) (periods []Period, err error) {
	doc, err := html.Parse(r)
	if err != nil {
		err = fmt.Errorf("failed to parse HTML: %w", err)
		return
	}

	tbody := findNodeWithTagName(doc, "tbody")

	if tbody == nil {
		err = fmt.Errorf("could not find the table listing the periods")
		return
	}

	rowIndex := 0

	const (
		columnActions = 3
		columnStatus  = 0
		columnStart   = 1
		columnEnd     = 2
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

		if len(cells) < 4 {
			continue
		}

		var period Period

		period.ID = extractIDFromActionsCell(cells[columnActions])

		period.Status, err = extractStatusFromStatusCell(cells[columnStatus])
		if err != nil {
			err = fmt.Errorf("row %d: %w", rowIndex, err)
			return
		}

		startStr := extractTextContent(cells[columnStart])
		period.Start, err = time.Parse(DateLayout, startStr)
		if err != nil {
			err = fmt.Errorf("row %d: failed to parse start time '%s': %s", rowIndex, startStr, err)
			return
		}

		endStr := extractTextContent(cells[columnEnd])
		period.End, err = time.Parse(DateLayout, endStr)
		if err != nil {
			err = fmt.Errorf("row %d: failed to parse end time '%s': %s", rowIndex, endStr, err)
			return
		}

		periods = append(periods, period)
	}
	return
}
