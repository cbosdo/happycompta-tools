// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"

	"log"

	"github.com/cbosdo/happycompta-tools/lib"
)

// loadImpl is the main logic entry point of the tool.
func loadImpl(cfg Config) error {
	client, err := lib.NewClient()
	if err != nil {
		return err
	}
	if err := client.Login(cfg.Email, cfg.Password); err != nil {
		return err
	}

	accounts, err := client.ListAccounts()
	if err != nil {
		return err
	}
	if len(accounts) == 0 {
		return errors.New("no bank account defined in happy-compta")
	}

	categories, err := client.ListCategories()
	if err != nil {
		return err
	}

	employees, err := client.ListEmployees()
	if err != nil {
		return err
	}

	providers, err := client.ListProviders()
	if err != nil {
		return err
	}

	periods, err := client.ListPeriods()
	if err != nil {
		return err
	}
	if len(periods) == 0 {
		return errors.New("no accounting period defined in happy-compta")
	}

	r, cleaner, err := getCSVReader(cfg.CSVPath, cfg.CSV)
	defer cleaner()
	if err != nil {
		return err
	}

	entries, err := parseCSV(r, cfg.CSV.Columns, cfg.Defaults, accounts, categories, employees, providers, periods)
	if err != nil {
		return err
	}

	// Add the receipts to the entries
	if err := addReceipts(cfg.Receipts, entries); err != nil {
		return err
	}

	// Load the entries to happy-compta
	for i, entry := range entries {
		err := client.AddEntry(&entry)
		if err != nil {
			log.Printf("failed to add entry #%d: %s", i, err)
		}

	}
	return nil
}
