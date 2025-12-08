// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/cbosdo/happycompta-tools/internal/common"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// toPain001 converts a CSV file to pain 001.001.03 for money transfers.
func toPain001(flags Config, dataPath string) error {
	// Read the CSV file
	reader, cleaner, err := common.GetCSVReader(flags.CSV.CSVParams, dataPath)
	if err != nil {
		return fmt.Errorf("failed to read CSV: %s", err)
	}
	defer cleaner()

	flags.Debtor.BIC = strings.ReplaceAll(flags.Debtor.BIC, " ", "")
	flags.Debtor.IBAN = strings.ReplaceAll(flags.Debtor.IBAN, " ", "")

	transferInit := NewTransferInitiation(flags.BatchID, &flags.Debtor)
	payment := Payment{}
	var header map[string]int
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error parsing the CSV file: %s", err)
		}

		if len(header) == 0 {
			header, err = getCSVHeader(flags.CSV.Columns, record)
			if err != nil {
				return err
			}
			continue
		}

		// Store the data
		amountStr := strings.ReplaceAll(record[header[columnsAmount]], "€", "")
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return fmt.Errorf("failed to parse amount %s to a number: %s", amountStr, err)
		}
		transaction := Transaction{
			Amount:     amount,
			Info:       sanitizeString(record[header[columnInfo]], 35),
			EndToEndID: sanitizeString(record[header[columnID]], 35),
			Creditor: Party{
				Name: sanitizeString(record[header[columnCreditor]], 140),
				IBAN: sanitizeID(record[header[columnIBAN]]),
				BIC:  sanitizeID(record[header[columnBIC]]),
			},
			Purpose: "REFU", // TODO Use an optional column for this
		}
		payment.Transactions = append(payment.Transactions, &transaction)
	}
	transferInit.AddPayment(&payment)

	// Write the pain001 file
	wr, cleaner, err := getOutputWriter(flags)
	defer cleaner()
	if err != nil {
		return err
	}
	return transferInit.Write(wr)
}

const (
	columnCreditor = "Creditor"
	columnIBAN     = "IBAN"
	columnBIC      = "BIC"
	columnID       = "EndToEndID"
	columnInfo     = "Info"
	columnsAmount  = "Amount"
)

func getCSVHeader(flags ColumnsConfig, record []string) (map[string]int, error) {
	var header = make(map[string]int)

	columns := []string{columnCreditor, columnIBAN, columnBIC, columnID, columnInfo, columnsAmount}
	flagsValue := reflect.ValueOf(flags)
	for _, column := range columns {
		csvName := flagsValue.FieldByName(column).String()
		idx := slices.Index(record, csvName)
		if idx < 0 {
			return header, fmt.Errorf("column not found in CSV file: %s", csvName)
		}
		header[column] = idx
	}

	return header, nil
}

func getOutputWriter(flags Config) (io.Writer, func(), error) {
	if flags.Output == "" {
		return os.Stdout, func() {}, nil
	}
	f, err := os.Create(flags.Output)
	if err != nil {
		return nil, func() {}, err
	}
	return f, func() { _ = f.Close() }, nil
}

// non breaking spaces and friends are hard to spot: replace them all!
var whitespaces = regexp.MustCompile(`[\p{Zs}]+`)

func sanitizeID(id string) string {
	return whitespaces.ReplaceAllString(id, "")
}

var invalidString = regexp.MustCompile("[^a-zA-Z0-9/?:().,'+ -]")

func sanitizeString(in string, maxLen int) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, in)

	if invalidString.MatchString(result) {
		log.Fatalf("String can only contain unaccented letter, digits and /-?:().,'+: '%s'", result)
	}

	if len(result) > maxLen {
		log.Fatalf("String cannot contain more than %d characters: '%s'", maxLen, result)
	}
	return result
}
