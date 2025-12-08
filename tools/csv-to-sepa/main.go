// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/cbosdo/happycompta-tools/internal/common"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	Output  string
	Debtor  Party
	BatchID string
	CSV     CsvConfig
}

type CsvConfig struct {
	common.CSVParams `mapstructure:",squash"`
	Columns          ColumnsConfig
}

type ColumnsConfig struct {
	Creditor   string
	IBAN       string
	BIC        string
	EndToEndID string `mapstructure:"id"`
	Amount     string
	Info       string
}

var rootCmd = &cobra.Command{
	Use:   path.Base(os.Args[0]) + "path/to/data",
	Short: "Convert a CSV file to a SEPA transfer file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var flags Config
		if err := viper.Unmarshal(&flags); err != nil {
			return fmt.Errorf("failed to parse configuration: %s", err)
		}
		return toPain001(flags, args[0])
	},
}

func init() {
	rootCmd.Flags().String("output", "", "SEPA file to write to. Defaults to stdout")
	rootCmd.Flags().String("batchid", "", "Unique identifier of the transfer initiation")
	rootCmd.Flags().String("debtor-name", "", "Debtor name")
	rootCmd.Flags().String("debtor-iban", "", "Debtor IBAN")
	rootCmd.Flags().String("debtor-bic", "", "Debtor BIC")
	rootCmd.Flags().String("csv-columns-creditor", "creditor", "Name of the column for the creditor name")
	rootCmd.Flags().String("csv-columns-iban", "iban", "Name of the column for the creditor's IBAN")
	rootCmd.Flags().String("csv-columns-bic", "bic", "Name of the column for the creditor's BIC")
	rootCmd.Flags().String("csv-columns-id", "id", "Name of the column for the end to end id")
	rootCmd.Flags().String("csv-columns-info", "info", "Name of the column for the transaction information")
	rootCmd.Flags().String("csv-columns-amount", "amount", "Name of the column for the transaction amount in euro")

	// CSV Structure flags
	rootCmd.Flags().String("csv-comma", "", "CSV field separator character.")
	rootCmd.Flags().String("csv-comment", "", "CSV comment character.")

	cobra.OnInitialize(func() { common.InitConfig(rootCmd) })

	rootCmd.Flags().VisitAll(common.BindFlagsToViper)

	viper.SetEnvPrefix("CSV_SEPA")
	viper.AutomaticEnv()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
