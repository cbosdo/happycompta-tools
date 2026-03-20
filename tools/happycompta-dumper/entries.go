// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cbosdo/happycompta-tools/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newEntriesCmd() *cobra.Command {
	var entriesCmd = &cobra.Command{
		Use:   "entries",
		Short: "List entries details",
		RunE: func(cmd *cobra.Command, args []string) error {
			var cfg Config

			if err := viper.Unmarshal(&cfg); err != nil {
				return fmt.Errorf("error unmarshaling the configuration: %s", err)
			}

			if cfg.Email == "" {
				log.Fatalf("email parameter or config value is required\n")
			}
			if cfg.Password == "" {
				log.Fatalf("password parameter or config value is required\n")
			}

			// Actually do something
			return entries(cfg, args[0])
		},
	}
	// TODO Add flags to filter the entries

	return entriesCmd
}

func entries(cfg Config, periodID string) error {
	client, err := lib.NewClient()
	if err != nil {
		return err
	}
	if err := client.Login(cfg.Email, cfg.Password); err != nil {
		return err
	}

	// TODO implement me
	entries, err := client.ListEntries(periodID)
	if err != nil {
		return err
	}

	w := csv.NewWriter(os.Stdout)
	w.Write([]string{"Entry ID", "Date", "Kind", "Title", "Amount", "Receipts"})
	for _, entry := range entries {
		amount := 0.0
		for _, allocation := range entry.Allocation {
			amount += allocation.Amount
		}
		w.Write([]string{
			entry.ID,
			entry.Date.Format("02-01-2006"),
			entry.Kind.String(),
			entry.Name,
			fmt.Sprintf("%f", amount),
			strings.Join(entry.Receipts, " "),
		})
	}
	w.Flush()
	return nil
}
