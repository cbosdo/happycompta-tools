// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/cbosdo/happycompta-tools/lib"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// parseCSV builds entries out of the CSV reader..
// Only the data from the CSV file are loaded, so no receipt will be attached by this function.
func parseCSV(
	r *csv.Reader,
	columnsCfg CSVColumns,
	defaults Defaults,
	accounts []lib.Account,
	categories []lib.Category,
	employees []lib.Employee,
	providers []lib.Provider,
	periods []lib.Period,
) (entries []lib.Entry, err error) {
	// Read the header and build the column map
	header, err := r.Read()
	if err == io.EOF {
		return nil, fmt.Errorf("CSV file is empty")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %s", err)
	}

	colMap := buildColumnMap(header, columnsCfg)
	log.Printf("CSV header read. Mapped columns: %+v", colMap)

	// Create maps for more efficient lookup later
	categoriesMap := createCategoriesMap(categories)
	employeesMap := createEmployeesMap(employees)
	providersMap := createProvidersMap(providers)
	periodsMap := createPeriodsMap(periods)

	var allErrors []error

	// Load each row as an entry
	for rowIndex := 1; ; rowIndex++ {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			allErrors = append(allErrors, fmt.Errorf("failed to read row %d: %s", rowIndex, err))
			continue
		}

		entry, err := createEntryFromRow(
			row, colMap, defaults, rowIndex, accounts, categoriesMap, employeesMap, providersMap, periodsMap,
		)
		if err != nil {
			allErrors = append(allErrors, fmt.Errorf("failed to process entry on row %d: %s", rowIndex, err))
			continue
		}

		entries = append(entries, entry)
	}

	err = errors.Join(allErrors...)
	return
}

func createCategoriesMap(slice []lib.Category) map[string]lib.Category {
	categories := map[string]lib.Category{}
	for _, category := range slice {
		categories[fmt.Sprintf("%s|%s", &category.Budget, category.Name)] = category
	}

	return categories
}

func stripDiacritics(in string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, in)
	return result
}

// Maps Lastname Firstname to employees.
func createEmployeesMap(slice []lib.Employee) map[string]lib.Employee {
	employees := map[string]lib.Employee{}
	for _, employee := range slice {
		fullName := strings.ToLower(fmt.Sprintf("%s %s", employee.Lastname, employee.Firstname))
		employees[stripDiacritics(fullName)] = employee
	}
	return employees
}

// Maps the names to the providers.
func createProvidersMap(slice []lib.Provider) map[string]lib.Provider {
	providers := map[string]lib.Provider{}
	for _, provider := range slice {
		providers[strings.ToLower(provider.Name)] = provider
	}
	return providers
}

// Maps <Start>-<End> dates to the period.
// Also map the empty string to the corresponding period since there can only be one.
func createPeriodsMap(slice []lib.Period) map[string]lib.Period {
	periods := map[string]lib.Period{}
	for _, period := range slice {
		periods[fmt.Sprintf("%s-%s", period.Start.Format(lib.DateLayout), period.End.Format(lib.DateLayout))] = period
		if period.Status == lib.PeriodStatusCurrent {
			periods[""] = period
		}
	}
	return periods
}

// Map column names from config to their index in the CSV file
type columnMap struct {
	Name     int
	Date     int
	Amount   int
	Stock    int
	Category int
	Comment  int
	Payment  int
	Budget   int
	Employee int
	Provider int
	Kind     int
	Period   int
	Bank     int
}

// buildColumnMap reads the header and maps the configured column names (e.g., cfg.Columns.Name)
// to their corresponding zero-based index in the CSV file.
func buildColumnMap(header []string, columns CSVColumns) columnMap {
	result := columnMap{
		Name:     -1,
		Date:     -1,
		Amount:   -1,
		Stock:    -1,
		Category: -1,
		Comment:  -1,
		Payment:  -1,
		Budget:   -1,
		Employee: -1,
		Provider: -1,
		Kind:     -1,
		Period:   -1,
		Bank:     -1,
	}

	colMap := map[string]*int{
		columns.Name:     &result.Name,
		columns.Date:     &result.Date,
		columns.Amount:   &result.Amount,
		columns.Stock:    &result.Stock,
		columns.Category: &result.Category,
		columns.Comment:  &result.Comment,
		columns.Payment:  &result.Payment,
		columns.Budget:   &result.Budget,
		columns.Employee: &result.Employee,
		columns.Provider: &result.Provider,
		columns.Kind:     &result.Kind,
		columns.Period:   &result.Period,
		columns.Bank:     &result.Bank,
	}

	for i, headerName := range header {
		if idxPtr, found := colMap[headerName]; found && headerName != "" {
			*idxPtr = i
		}
	}
	return result
}

// getField safely retrieves a field value from the row slice.
func getField(row []string, colIndex int) string {
	if colIndex >= 0 && colIndex < len(row) {
		return strings.TrimSpace(row[colIndex])
	}
	return ""
}

// getOptionalField retrieves a field value from the row slice, falling back to a default if the field is empty.
func getOptionalField(row []string, colIndex int, defaultValue string) string {
	value := getField(row, colIndex)
	if value == "" {
		return defaultValue
	}
	return value
}

// createEntryFromRow processes a single CSV row and maps it to a lib.Entry.
func createEntryFromRow(
	row []string,
	colMap columnMap,
	defaults Defaults,
	rowIndex int,
	accounts []lib.Account,
	categories map[string]lib.Category,
	employees map[string]lib.Employee,
	providers map[string]lib.Provider,
	periods map[string]lib.Period,
) (entry lib.Entry, err error) {
	var allErrors []error // Initialize a slice to collect errors

	// Date
	dateStr := getField(row, colMap.Date)
	if dateStr == "" {
		allErrors = append(allErrors, fmt.Errorf("date column is missing or empty"))
	} else {
		date, dateErr := time.Parse(lib.DateLayout, dateStr)
		if dateErr != nil {
			allErrors = append(allErrors, fmt.Errorf("failed to parse date '%s': %w", dateStr, dateErr))
		} else {
			entry.Date = date
		}
	}

	// Name
	entry.Name = getField(row, colMap.Name)

	// Amount. May not be needed for checks allocations
	amountStr := getField(row, colMap.Amount)
	amount := 0.0
	if amountStr != "" {
		var amountErr error
		amount, amountErr = parseAmount(amountStr)
		if amountErr != nil {
			allErrors = append(allErrors, fmt.Errorf("failed to parse amount '%s': %s", amountStr, amountErr))
		}
	}

	// Comment
	entry.Comment = getField(row, colMap.Comment)

	// Kind
	kind := getOptionalField(row, colMap.Kind, defaults.Kind)
	entry.Kind = lib.NewKind(kind)
	if entry.Kind == lib.KindUndefined {
		allErrors = append(allErrors, fmt.Errorf(
			"invalid entry type '%s', accepted values are %s, %s and %s",
			kind, lib.KindSpend, lib.KindTake, lib.KindAllocation,
		))
	}

	// Budget, the accepted values are FON, ASC or AEP.
	budgetStr := getOptionalField(row, colMap.Budget, defaults.Budget)
	if budgetStr != "" {
		entry.Budget = lib.NewBudgetFromString(budgetStr)
	}
	if entry.Budget == lib.BudgetUndefined {
		allErrors = append(allErrors, fmt.Errorf("invalid budget '%s'", budgetStr))
	}

	// PaymentMethod
	paymentMethodStr := getOptionalField(row, colMap.Payment, defaults.Payment)
	if paymentMethodStr != "" {
		paymentMethod := lib.NewPaymentMethodFromString(paymentMethodStr)
		if paymentMethod != lib.PaymentMethodUndefined {
			entry.PaymentMethod = paymentMethod
		} else {
			allErrors = append(allErrors, fmt.Errorf("invalid payment method '%s'", paymentMethodStr))
		}
	} else {
		allErrors = append(allErrors, fmt.Errorf("missing payment method"))
	}

	// Category
	categoryName := getOptionalField(row, colMap.Category, defaults.Category)
	var category lib.Category
	categoryOK := false

	// Only attempt category lookup if budget is valid (to avoid logging redundant errors)
	if entry.Budget != lib.BudgetUndefined {
		categoryKey := fmt.Sprintf("%s|%s", entry.Budget, categoryName)
		category, categoryOK = categories[categoryKey]

		if !categoryOK {
			allErrors = append(allErrors, fmt.Errorf(
				"invalid category '%s' name / '%s' budget combination",
				categoryName, entry.Budget,
			))
		}
	}

	// Stock (Only check if category lookup was successful)
	stock := 0
	if categoryOK && bool(category.Stock) {
		stockStr := getField(row, colMap.Stock)
		if stockStr == "" {
			allErrors = append(allErrors, fmt.Errorf("no stock defined but %s category needs it", category.Name))
		} else {
			var stockErr error
			stock, stockErr = strconv.Atoi(stockStr)
			if stockErr != nil {
				allErrors = append(allErrors, fmt.Errorf("failed to parse '%s' stock as an integer", stockStr))
			}
		}
	} else if amountStr == "" {
		allErrors = append(allErrors, fmt.Errorf("missing required amount value for row %d", rowIndex))
	}

	entry.Allocation = []lib.AllocationLine{
		{
			CategoryID: category.ID,
			Amount:     amount,
			Stock:      stock,
		},
	}

	// Party: the employee and provider fields are mutually exclusive and optional.
	employeeStr := getField(row, colMap.Employee)
	providerStr := getField(row, colMap.Provider)
	if employeeStr != "" && providerStr != "" {
		allErrors = append(allErrors, fmt.Errorf("has both employee ('%s') and provider ('%s') specified", employeeStr, providerStr))
	} else {
		if employeeStr != "" {
			employee, ok := employees[stripDiacritics(strings.ToLower(employeeStr))]
			if !ok {
				allErrors = append(allErrors, fmt.Errorf(
					"unknown employee '%s', the value needs to be in the <Lastname> <Firstname> format",
					employeeStr,
				))
			} else {
				entry.Party = &employee
			}
		}

		if providerStr != "" {
			provider, ok := providers[strings.ToLower(providerStr)]
			if !ok {
				allErrors = append(allErrors, fmt.Errorf(
					"unknown provider '%s', the value needs to match the name of an existing provider",
					providerStr,
				))
			} else {
				entry.Party = &provider
			}
		}
	}

	// Look for the period
	periodStr := getField(row, colMap.Period)
	if periodStr == "" {
		periodStr = defaults.Period
	}
	period, ok := periods[periodStr]
	if !ok {
		allErrors = append(allErrors, fmt.Errorf("couldn't find the '%s' period. Is there a current one defined?", periodStr))
	} else {
		entry.Period = period.ID
	}

	// Look for the account
	bank := getOptionalField(row, colMap.Bank, defaults.Bank)
	// Only try to get account if the budget was successfully determined
	if entry.Budget != lib.BudgetUndefined {
		account, accErr := getAccountFromBankBudget(accounts, bank, entry.Budget)
		if accErr != nil {
			allErrors = append(allErrors, fmt.Errorf("failed to find account: %w", accErr))
		} else {
			entry.Account = account
		}
	}

	// Check for collected errors
	if len(allErrors) > 0 {
		// Return an empty entry and the combined errors
		return lib.Entry{}, errors.Join(allErrors...)
	}

	return entry, nil
}

func getAccountFromBankBudget(
	accounts []lib.Account, bank string, budget lib.Budget,
) (result lib.Account, err error) {
	banks := []string{}
	for _, account := range accounts {
		if !slices.Contains(banks, account.Bank) {
			banks = append(banks, account.Bank)
		}
	}
	if bank == "" {
		if len(banks) > 1 {
			err = errors.New("more than one bank found, you have to provide the name of the bank holding the account")
			return
		}
		// Using the only bank that we found by default
		bank = banks[0]
	}

	matchingAllBudgets := []lib.Account{}
	matching := []lib.Account{}
	for _, account := range accounts {
		if strings.EqualFold(account.Bank, bank) {
			switch account.Budget {
			case budget:
				matching = append(matching, account)
			case lib.BudgetUndefined:
				// Undefined budget on an account means both ASC and FON
				matchingAllBudgets = append(matchingAllBudgets, account)
			}
		}
	}

	// We may have found more than one account.
	// The common situation would be: 1 with the expected budget and 1 with both.
	// I don't think anything on happy-compta prevents from having more than one account for the same budget in the
	// same bank, but this is rather unlikely to happen.
	if len(matching) == 1 {
		result = matching[0]
		return
	} else if len(matching) > 1 {
		err = fmt.Errorf(
			"more than one account found for the %s budget at %s bank. This is not supported yet",
			budget.String(), bank,
		)
		return
	} else if len(matchingAllBudgets) == 1 {
		result = matchingAllBudgets[0]
		return
	} else if len(matchingAllBudgets) > 1 {
		err = fmt.Errorf(
			"more than one account found for the both budgets at %s bank. This is not supported yet", bank,
		)
		return
	}

	err = fmt.Errorf("no account found matching the %s budget at %s bank", budget.String(), bank)
	return
}
