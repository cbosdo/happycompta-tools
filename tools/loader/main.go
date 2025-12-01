// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var load func(Config) error = loadImpl

// Define the root command
var rootCmd = &cobra.Command{
	Use:   "loader path/to/file.csv",
	Short: "A program loading entries from a CSV file as entries into happy-compta",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var cfg Config

		if err := viper.Unmarshal(&cfg); err != nil {
			return fmt.Errorf("Error unmarshaling the configuration: %s", err)
		}
		cfg.CSVPath = args[0]

		if cfg.Email == "" {
			log.Fatalf("Email parameter or config value is required\n")
		}
		if cfg.Password == "" {
			log.Fatalf("Password parameter or config value is required\n")
		}

		// Actually do something
		return load(cfg)
	},
}

func init() {
	rootCmd.PersistentFlags().StringP("config", "c", "", "Configuration file path")
	rootCmd.PersistentFlags().String("email", "", "User email address (REQUIRED)")
	rootCmd.PersistentFlags().String("password", "", "User password (REQUIRED)")

	rootCmd.Flags().String("receipts", "receipts", "Folder containing the receipts")
	// TODO We probably need to add a parameter to select the way the receipts are matched

	// Default Value flags
	rootCmd.Flags().String("budget", "", "Default value for budget column.")
	rootCmd.Flags().String("bank", "", "Default value for bank column.")
	rootCmd.Flags().String("category", "", "Default value for category column.")
	rootCmd.Flags().String("payment", "", "Default value for payment column.")
	rootCmd.Flags().String("kind", "", "Default value for kind column.")
	rootCmd.Flags().String("period", "", "Accounting period to add the entries to. Defaults to the current one.")

	// CSV Structure flags
	rootCmd.Flags().String("csv-comma", "", "CSV field separator character.")
	rootCmd.Flags().String("csv-comment", "", "CSV comment character.")

	// CSV Column mapping flags
	rootCmd.Flags().String("csv-columns-name", "name", "CSV column name for transaction name.")
	rootCmd.Flags().String("csv-columns-date", "date", "CSV column name for date.")
	rootCmd.Flags().String("csv-columns-amount", "amount", "CSV column name for amount.")
	rootCmd.Flags().String("csv-columns-stock", "amount", `CSV column name for the stock.
This is usually needed for check allocations and orders.`)
	rootCmd.Flags().String("csv-columns-category", "category", "CSV column name for category.")
	rootCmd.Flags().String("csv-columns-comment", "comment", "CSV column name for comment.")
	rootCmd.Flags().String("csv-columns-payment", "payment", "CSV column name for payment type.")
	rootCmd.Flags().String("csv-columns-budget", "budget", "CSV column name for budget ID.")
	rootCmd.Flags().String("csv-columns-employee", "employee", "CSV column name for employee.")
	rootCmd.Flags().String("csv-columns-provider", "provider", "CSV column name for provider.")
	rootCmd.Flags().String("csv-columns-period", "period", "CSV column name for the period.")
	rootCmd.Flags().String("csv-columns-bank", "account", `CSV column name for the name of the bank holding the account.
This is used in conjunction with the budget to identify the target account.`)

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().VisitAll(bindFlagsToViper)
	rootCmd.Flags().VisitAll(bindFlagsToViper)

	viper.SetEnvPrefix("LOADER")
	viper.AutomaticEnv()
}

// bindFlagsToViper is a helper function to bind a flag to a Viper key.
func bindFlagsToViper(flag *pflag.Flag) {
	key := strings.ReplaceAll(flag.Name, "-", ".")

	if flag.Name == "config" {
		return
	}

	if err := viper.BindPFlag(key, flag); err != nil {
		log.Fatalf("error binding flag '%s' to viper key '%s': %v\n", flag.Name, key, err)
	}
}

func initConfig() {
	configPath, err := rootCmd.PersistentFlags().GetString("config")
	if err != nil {
		log.Fatalf("Error reading config flag: %s", err)
	}

	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok && configPath == "" {
			return
		}

		log.Fatalf("Error loading configuration: %s\n", err)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
