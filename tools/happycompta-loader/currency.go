// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// parseAmount reads a currency in either US or European format into a float.
func parseAmount(input string) (float64, error) {
	if input == "" {
		return 0, errors.New("amount is missing or empty")
	}

	const usCurrencyPattern = `^€?\s?(\d{1,3}(,\d{3})*|\d+)(\.\d{2})?\s?€?$`
	var usCurrencyRegex = regexp.MustCompile(usCurrencyPattern)

	cleanInput := input
	// We only handle Euros for now since happy-compta doesn't handle any other currency.
	cleanInput = strings.ReplaceAll(cleanInput, "€", "")
	if usCurrencyRegex.MatchString(input) {
		cleanInput = strings.ReplaceAll(cleanInput, ",", "")
		cleanInput = strings.TrimSpace(cleanInput)
	} else {
		cleanInput = strings.ReplaceAll(cleanInput, ".", "")      // Remove dots
		cleanInput = strings.ReplaceAll(cleanInput, " ", "")      // Remove regular spaces
		cleanInput = strings.ReplaceAll(cleanInput, "\u00A0", "") // Remove non-breaking spaces
		cleanInput = strings.ReplaceAll(cleanInput, ",", ".")     // Change decimal comma to dot
	}

	amount, err := strconv.ParseFloat(cleanInput, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse amount '%s' (cleaned: '%s'): %w", input, cleanInput, err)
	}

	return amount, nil
}
