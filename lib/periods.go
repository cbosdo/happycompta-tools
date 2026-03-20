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
	resp, err := c.client.Get(url_base + "/operations/index")
	if err != nil {
		err = fmt.Errorf("failed to get the operations page: %s", err)
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

// parsePeriods reads the periods from HTML content.
func parsePeriods(r io.Reader) (periods []Period, err error) {
	doc, err := html.Parse(r)
	if err != nil {
		err = fmt.Errorf("failed to parse HTML: %w", err)
		return
	}

	selectNode := findNodeWithKeyValueAttr(doc, "name", "exercice_id")
	if selectNode == nil {
		err = fmt.Errorf("could not find the select listing the periods")
		return
	}

	// Regex to extract dates and status text
	// Example: Du 01/01/2026 au 31/12/2026 [En cours]
	re := regexp.MustCompile(`Du (\d{2}/\d{2}/\d{4}) au (\d{2}/\d{2}/\d{4}) \[(.+)\]`)

	for c := selectNode.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "option" {
			var p Period
			p.ID = getAttr(c, "value")

			text := html.UnescapeString(extractTextContent(c))
			matches := re.FindStringSubmatch(text)

			if len(matches) == 4 {
				// Parse Start Date
				p.Start, err = time.Parse(DateLayout, matches[1])
				if err != nil {
					return nil, fmt.Errorf("failed to parse start date %s: %w", matches[1], err)
				}

				// Parse End Date
				p.End, err = time.Parse(DateLayout, matches[2])
				if err != nil {
					return nil, fmt.Errorf("failed to parse end date %s: %w", matches[2], err)
				}

				// Map status string to PeriodStatus type
				statusText := matches[3]
				switch statusText {
				case "En cours":
					p.Status = PeriodStatusCurrent
				case "Clôture provisoire":
					p.Status = PeriodStatusProvisionallyClosed
				case "Clôture définitive":
					p.Status = PeriodStatusDefinitelyClosed
				default:
					p.Status = PeriodStatusUndefined
				}

				periods = append(periods, p)
			}
		}
	}

	return
}
