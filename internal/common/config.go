// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// BindFlagsToViper is a helper function to bind a flag to a Viper key.
func BindFlagsToViper(flag *pflag.Flag) {
	key := strings.ReplaceAll(flag.Name, "-", ".")

	if flag.Name == "config" {
		return
	}

	if err := viper.BindPFlag(key, flag); err != nil {
		log.Fatalf("error binding flag '%s' to viper key '%s': %v\n", flag.Name, key, err)
	}
}

func InitConfig(cmd *cobra.Command) {
	configPath, err := cmd.PersistentFlags().GetString("config")
	if err != nil {
		log.Fatalf("error reading config flag: %s", err)
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

		log.Fatalf("error loading configuration: %s\n", err)
	}
}
