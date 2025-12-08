// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"

	"github.com/cbosdo/happycompta-tools/internal/common"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// These variables are set during the build process via ldflags.
var (
	version  = "dev"
	revision = "HEAD"
)

// Config holds the application parameters.
type Config struct {
	Email    string `mapstructure:"email"`
	Password string `mapstructure:"password"`
}

// Define the root command
var rootCmd = &cobra.Command{
	Use:     "dumper",
	Short:   "A program dumping data from happy-compta",
	Version: fmt.Sprintf("%s (%s)", version, revision),
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
		return dump(cfg)
	},
}

func init() {
	rootCmd.PersistentFlags().StringP("config", "c", "", "Configuration file path")
	rootCmd.PersistentFlags().String("email", "", "User email address (REQUIRED)")
	rootCmd.PersistentFlags().String("password", "", "User password (REQUIRED)")

	rootCmd.SetVersionTemplate("{{.Version}}\n")

	cobra.OnInitialize(func() { common.InitConfig(rootCmd) })

	rootCmd.PersistentFlags().VisitAll(common.BindFlagsToViper)
	rootCmd.Flags().VisitAll(common.BindFlagsToViper)

	viper.SetEnvPrefix("LOADER")
	viper.AutomaticEnv()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
