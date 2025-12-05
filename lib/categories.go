// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Category struct {
	ID       int
	ParentID int  `json:"parent_id"`
	Kind     Kind `json:"type"`
	Name     string
	Budget   Budget  `json:"section_id"`
	Stock    IntBool `json:"stock"`
}

// ListCategories gets all the operation categories defined for the organization.
func (c *Client) ListCategories() (categories []Category, err error) {
	resp, err := c.client.Get(url_base + "/ajax/get-categories")
	if err != nil {
		err = fmt.Errorf("failed to get the categories: %s", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to get the categories, got %d status code", resp.StatusCode)
		return
	}

	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(&categories); err != nil {
		err = fmt.Errorf("failed to parse categories data: %s", err)
		return
	}
	return
}
