// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/cbosdo/happycompta-tools/internal/common"
)

// CSVColumns holds the mapping for individual column names in the CSV file.
type CSVColumns struct {
	Name     string `mapstructure:"name"`
	Date     string `mapstructure:"date"`
	Amount   string `mapstructure:"amount"`
	Stock    string `mapstructure:"stock"`
	Category string `mapstructure:"category"`
	Comment  string `mapstructure:"comment"`
	Payment  string `mapstructure:"payment"`
	Budget   string `mapstructure:"budget"`
	Employee string `mapstructure:"employee"`
	Provider string `mapstructure:"provider"`
	Kind     string `mapstructure:"kind"`
	Period   string `mapstructure:"period"`
	Bank     string `mapstructure:"bank"`
}

// CSVConfig provides a logical grouping for all CSV-related settings.
type CSVConfig struct {
	common.CSVParams `mapstructure:",squash"`
	Columns          CSVColumns `mapstructure:"columns"`
}

// getSingleRune converts a string field to a rune, validating that it's a single character.
// If the field is empty, it returns 0.
func getSingleRune(value, fieldName string) (rune, error) {
	if value == "" {
		return 0, nil
	}
	runes := []rune(value)
	if len(runes) != 1 {
		return 0, fmt.Errorf("%s must be a single character, but got '%s'", fieldName, value)
	}
	return runes[0], nil
}

// GetCommaRune converts the Commma string field to a rune, validating that it's a single character.
// If the field is empty, it returns 0, allowing the csv.Reader to use its default.
func (c *CSVConfig) GetCommaRune() (rune, error) {
	return getSingleRune(c.Comma, "comma separator")
}

// GetCommentRune converts the Comment string field to a rune, validating that it's a single character.
// If the field is empty, it returns 0, allowing the csv.Reader to use its default.
func (c *CSVConfig) GetCommentRune() (rune, error) {
	return getSingleRune(c.Comment, "comment character")
}

// Defaults holds the default values for optional columns.
type Defaults struct {
	Budget   string `mapstructure:"budget"`
	Bank     string `mapstructure:"bank"`
	Category string `mapstructure:"category"`
	Payment  string `mapstructure:"payment"`
	Kind     string `mapstructure:"kind"`
	Period   string `mapstructure:"period"`
}

// Config holds the application parameters.
type Config struct {
	Email    string    `mapstructure:"email"`
	Password string    `mapstructure:"password"`
	Receipts string    `mapstructure:"receipts"`
	CSV      CSVConfig `mapstructure:"csv"`
	CSVPath  string
	Defaults Defaults `mapstructure:",squash"`
}
