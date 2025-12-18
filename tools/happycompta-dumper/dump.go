// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/cbosdo/happycompta-tools/lib"
)

func dump(cfg Config) error {
	fmt.Printf("Dump happy-compta data for test purpose\n")

	client, err := lib.NewClient()
	if err != nil {
		return err
	}
	if err := client.Login(cfg.Email, cfg.Password); err != nil {
		return err
	}

	employees, err := client.ListEmployees()
	if err != nil {
		return err
	}

	fmt.Printf("Employees (%d):\n", len(employees))
	for _, emp := range employees {
		active := "inactive"
		if emp.Active {
			active = "active"
		}

		fmt.Printf("%s: %s,%s (%s)\n", emp.ID, emp.Lastname, emp.Firstname, active)
	}

	providers, err := client.ListProviders()
	if err != nil {
		return err
	}
	fmt.Printf("\nProviders (%d):\n", len(providers))
	for _, p := range providers {
		archived := ""
		if p.Archived {
			archived = " (Archived)"
		}
		fmt.Printf(
			"%s: %s%s\n    %s - %s %s\n    %s\n    %s\n    %s\n",
			p.ID, p.Name, archived,
			p.Address, p.ZipCode, p.City,
			p.Phone,
			p.Email,
			p.Comment,
		)
	}

	periods, err := client.ListPeriods()
	fmt.Printf("\nPeriods:\n")
	if err != nil {
		return err
	}
	for _, p := range periods {
		fmt.Printf("%s: %s - %s (%d)\n", p.ID, p.Start.Format(lib.DateLayout), p.End.Format(lib.DateLayout), p.Status)
	}

	accounts, err := client.ListAccounts()
	if err != nil {
		return err
	}
	fmt.Printf("\nAccounts:\n")
	for _, account := range accounts {
		fmt.Printf("%d: %s (%d - %s)\n", account.ID, account.Bank, account.Budget, account.Abbrev)
	}

	categories, err := client.ListCategories()
	if err != nil {
		return err
	}
	fmt.Printf("\nCategories (%d)\n", len(categories))
	for _, category := range categories {
		fmt.Printf(
			"%d: %s (%s), parent: %d, section: %d\n",
			category.ID,
			category.Name,
			category.Kind,
			category.ParentID,
			category.Budget,
		)
	}
	return nil
}
