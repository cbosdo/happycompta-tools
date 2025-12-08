// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"encoding/csv"
	"fmt"
	"os"
)

// CSVParams holds the configuration for the CSV reader's low-level parameters.
type CSVParams struct {
	Comma   string `mapstructure:"comma"`
	Comment string `mapstructure:"comment"`
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

// GetCommaRune converts the Commma string field to a rune.
func (c *CSVParams) GetCommaRune() (rune, error) {
	return getSingleRune(c.Comma, "comma separator")
}

// GetCommentRune converts the Comment string field to a rune.
func (c *CSVParams) GetCommentRune() (rune, error) {
	return getSingleRune(c.Comment, "comment character")
}

// GetCSVReader opens the file, creates a csv.Reader, and applies the given CSV configuration parameters.
// The returned cleaner function must be called when the reader is no longer needed.
func GetCSVReader(params CSVParams, dataPath string) (*csv.Reader, func(), error) {
	file, err := os.Open(dataPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open CSV file %s: %w", dataPath, err)
	}
	cleaner := func() { _ = file.Close() }

	r := csv.NewReader(file)

	commaRune, err := params.GetCommaRune()
	if err != nil {
		cleaner()
		return nil, nil, fmt.Errorf("CSV comma config error: %w", err)
	}
	if commaRune != 0 {
		r.Comma = commaRune
	}

	commentRune, err := params.GetCommentRune()
	if err != nil {
		cleaner()
		return nil, nil, fmt.Errorf("CSV comment config error: %w", err)
	}
	if commentRune != 0 {
		r.Comment = commentRune
	}

	return r, cleaner, nil
}
