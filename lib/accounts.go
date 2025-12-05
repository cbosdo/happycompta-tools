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

// Account represent a bank account of the organization.
type Account struct {
	ID     int
	Bank   string `json:"banque"`
	Budget Budget `json:"type"`
	Abbrev string `json:"abreviation"`
}

// ListAccounts lists all the bank accounts of the organization.
func (c *Client) ListAccounts() (accounts []Account, err error) {
	resp, err := c.client.Get(url_base + "/ajax/get-comptes")
	if err != nil {
		err = fmt.Errorf("failed to get the accounts: %s", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to get the accounts, got %d status code", resp.StatusCode)
		return
	}

	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(&accounts); err != nil {
		err = fmt.Errorf("failed to parse accounts data: %s", err)
		return
	}
	return
}
