// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/csv"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/cbosdo/happycompta-tools/lib"
)

func TestBuildColumnMap(t *testing.T) {
	configCols := CSVColumns{
		Name:     "Transaction_Name",
		Date:     "Date_Of_Tx",
		Amount:   "Tx_Amount",
		Category: "Type_Category",
		Budget:   "Budget_Code",
		Provider: "Vendor",
		Employee: "Employee",
	}

	tests := []struct {
		name    string
		header  []string
		config  CSVColumns
		wantMap columnMap
	}{
		{
			name:   "Full match and correct ordering",
			header: []string{"Date_Of_Tx", "Transaction_Name", "Tx_Amount", "Budget_Code", "Vendor"},
			config: configCols,
			wantMap: columnMap{
				Date:     0,
				Name:     1,
				Amount:   2,
				Budget:   3,
				Provider: 4,
				Category: -1, // Not present
				Comment:  -1,
				Payment:  -1,
				Employee: -1,
				Kind:     -1,
				Period:   -1,
				Stock:    -1,
				Bank:     -1,
			},
		},
		{
			name:   "Partial match and missing columns",
			header: []string{"Transaction_Name", "Tx_Amount"},
			config: configCols,
			wantMap: columnMap{
				Name:     0,
				Amount:   1,
				Date:     -1,
				Category: -1,
				Comment:  -1,
				Payment:  -1,
				Budget:   -1,
				Employee: -1,
				Provider: -1,
				Kind:     -1,
				Period:   -1,
				Stock:    -1,
				Bank:     -1,
			},
		},
		{
			name:   "Header reordering",
			header: []string{"Budget_Code", "Date_Of_Tx", "Transaction_Name"},
			config: configCols,
			wantMap: columnMap{
				Budget:   0,
				Date:     1,
				Name:     2,
				Amount:   -1,
				Category: -1,
				Comment:  -1,
				Payment:  -1,
				Employee: -1,
				Provider: -1,
				Kind:     -1,
				Period:   -1,
				Stock:    -1,
				Bank:     -1,
			},
		},
		{
			name:   "Empty header",
			header: []string{},
			config: configCols,
			wantMap: columnMap{
				Name:     -1,
				Date:     -1,
				Amount:   -1,
				Category: -1,
				Comment:  -1,
				Payment:  -1,
				Budget:   -1,
				Employee: -1,
				Provider: -1,
				Kind:     -1,
				Period:   -1,
				Stock:    -1,
				Bank:     -1,
			},
		},
		{
			name:   "Configured name is empty",
			header: []string{"Transaction_Name"},
			config: CSVColumns{Name: "", Date: "Date_Of_Tx"}, // Name config is empty
			wantMap: columnMap{
				Name:     -1, // Should not map because the header 'Transaction_Name' doesn't match the empty config value
				Date:     -1,
				Amount:   -1,
				Category: -1,
				Comment:  -1,
				Payment:  -1,
				Budget:   -1,
				Employee: -1,
				Provider: -1,
				Kind:     -1,
				Period:   -1,
				Stock:    -1,
				Bank:     -1,
			},
		},
		{
			name:   "Header names need exact match",
			header: []string{" Date_Of_Tx ", "Transaction_Name"},
			config: configCols,
			wantMap: columnMap{
				Date:     -1, // CSVReader trims the header name, no need to do it
				Name:     1,
				Amount:   -1,
				Category: -1,
				Comment:  -1,
				Payment:  -1,
				Budget:   -1,
				Employee: -1,
				Provider: -1,
				Kind:     -1,
				Period:   -1,
				Stock:    -1,
				Bank:     -1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMap := buildColumnMap(tt.header, tt.config)
			if !reflect.DeepEqual(gotMap, tt.wantMap) {
				t.Errorf("buildColumnMap() got = %+v, want %+v", gotMap, tt.wantMap)
			}
		})
	}
}

// baseTime is a reusable time reference for tests
var baseTime = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

func getMockCategories() []lib.Category {
	return []lib.Category{
		{ID: 100, Name: "Office Supplies", Budget: lib.BudgetFON, Stock: false},
		{ID: 101, Name: "Rent", Budget: lib.BudgetFON, Stock: false},
		{ID: 200, Name: "Gifts", Budget: lib.BudgetASC, Stock: false},
		{ID: 201, Name: "Check Alloc", Budget: lib.BudgetASC, Stock: true}, // Requires stock
		{ID: 300, Name: "Unused", Budget: lib.BudgetFON},
	}
}

func getMockPeriods() []lib.Period {
	return []lib.Period{
		{
			ID:     "12345",
			Start:  baseTime,
			End:    time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
			Status: lib.PeriodStatusCurrent,
		},
		{
			ID:     "12346",
			Start:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			End:    time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
			Status: lib.PeriodStatusDefinitelyClosed,
		},
	}
}

// getBaseDefaults returns the common default configuration.
func getBaseDefaults() Defaults {
	return Defaults{
		Budget:   "FON",
		Category: "Office Supplies",
		Payment:  "card",
		Kind:     "depenses",
		Period:   "", // defaults to current period
	}
}

// getMinimalColMap returns the standard column mapping setup.
func getMinimalColMap() columnMap {
	return buildColumnMap(
		[]string{"DATE", "NAME", "AMOUNT", "CATEGORY", "BUDGET", "EMPLOYEE",
			"PROVIDER", "PAYMENT", "KIND", "COMMENT", "STOCK", "PERIOD", "BANK"},
		CSVColumns{
			Date:     "DATE",
			Name:     "NAME",
			Amount:   "AMOUNT",
			Category: "CATEGORY",
			Budget:   "BUDGET",
			Employee: "EMPLOYEE",
			Provider: "PROVIDER",
			Payment:  "PAYMENT",
			Kind:     "KIND",
			Comment:  "COMMENT",
			Stock:    "STOCK",
			Period:   "PERIOD",
			Bank:     "BANK",
		},
	)
}

func TestCreateEntryFromRow_Success(t *testing.T) {
	colMap := getMinimalColMap()
	accounts := []lib.Account{
		{ID: 10, Bank: "First National Bank", Budget: lib.BudgetFON, Abbrev: "FNB"},
	}
	defaults := getBaseDefaults()
	categoriesMap := createCategoriesMap(getMockCategories())
	employeesMap := createEmployeesMap([]lib.Employee{
		{ID: "E10", Lastname: "DOE", Firstname: "JOHN", Active: true},
	})
	providersMap := createProvidersMap([]lib.Provider{
		{ID: "P50", Name: "TechCorp Solutions", City: "Faketown"},
	})
	periodsMap := createPeriodsMap(getMockPeriods())

	// Expected base entry values
	expected := lib.Entry{
		Period:        "12345",
		Kind:          lib.KindSpend,
		Date:          baseTime,
		Name:          "Test Purchase",
		Budget:        lib.BudgetFON,
		PaymentMethod: lib.PaymentMethodCard,
		Account:       accounts[0],
		Comment:       "Test comment",
		Allocation: []lib.AllocationLine{
			{
				CategoryID: 100, // Office Supplies
				Amount:     100.50,
				Stock:      0,
			},
		},
	}

	// Row uses defaults for Budget, Category, Payment, Kind, Period
	row := []string{
		"01/01/2025",          // DATE
		"Test Purchase",       // NAME
		"100.50€",             // AMOUNT
		"",                    // CATEGORY (use default "Office Supplies")
		"",                    // BUDGET (use default "FON")
		"",                    // EMPLOYEE
		"",                    // PROVIDER
		"",                    // PAYMENT (use default "card")
		"",                    // KIND (use default "depenses")
		"Test comment",        // COMMENT
		"",                    // STOCK
		"",                    // PERIOD (use default "")
		"First National Bank", // BANK
	}

	entry, err := createEntryFromRow(row, colMap, defaults, 1, accounts,
		categoriesMap, employeesMap, providersMap, periodsMap)

	if err != nil {
		t.Fatalf("createEntryFromRow failed unexpectedly: %v", err)
	}

	if !reflect.DeepEqual(entry, expected) {
		t.Errorf("Entry mismatch.\nGot:  %+v\nWant: %+v", entry, expected)
	}
}

func TestCreateEntryFromRow_PartyMutualExclusion(t *testing.T) {
	colMap := getMinimalColMap()
	accounts := []lib.Account{
		{ID: 10, Bank: "First National Bank", Budget: lib.BudgetFON, Abbrev: "FNB"},
	}
	defaults := getBaseDefaults()
	categoriesMap := createCategoriesMap(getMockCategories())
	employeesMap := createEmployeesMap([]lib.Employee{
		{ID: "E10", Lastname: "DOE", Firstname: "JOHN", Active: true},
	})
	providersMap := createProvidersMap([]lib.Provider{
		{ID: "P50", Name: "TechCorp Solutions", City: "Faketown"},
	})
	periodsMap := createPeriodsMap(getMockPeriods())

	// Row specifying BOTH Employee and Provider (ERROR case)
	row := []string{
		"01/01/2025", "Test", "10", "Office Supplies", "FON", "John Doe",
		"TechCorp Solutions", "card", "depenses", "", "", "", "First National Bank",
	}

	_, err := createEntryFromRow(row, colMap, defaults, 1, accounts,
		categoriesMap, employeesMap, providersMap, periodsMap)

	if err == nil || !strings.Contains(err.Error(), "has both employee") {
		t.Errorf("Expected mutual exclusion error, got: %v", err)
	}
}

func TestCreateEntryFromRow_StockRequired(t *testing.T) {
	colMap := getMinimalColMap()
	accounts := []lib.Account{
		{ID: 20, Bank: "Global Reserve", Budget: lib.BudgetASC, Abbrev: "GR"},
	}
	defaults := getBaseDefaults()
	categoriesMap := createCategoriesMap(getMockCategories())
	employeesMap := createEmployeesMap([]lib.Employee{
		{ID: "E10", Lastname: "DOE", Firstname: "JOHN", Active: true},
	})
	providersMap := createProvidersMap([]lib.Provider{
		{ID: "P50", Name: "TechCorp Solutions", City: "Faketown"},
	})
	periodsMap := createPeriodsMap(getMockPeriods())

	// Row using Category "Check Alloc" which requires stock, but leaving stock column empty.
	row := []string{
		"01/01/2025", "Test", "10", "Check Alloc", "ASC", "", "",
		"check allocation", "attributions", "", "", "", "Global Reserve",
	}

	_, err := createEntryFromRow(row, colMap, defaults, 1, accounts,
		categoriesMap, employeesMap, providersMap, periodsMap)

	if err == nil || !strings.Contains(err.Error(), "no stock defined") {
		t.Errorf("Expected 'no stock defined' error, got: %v", err)
	}
}

func TestCreateEntryFromRow_DateParsingFailure(t *testing.T) {
	colMap := getMinimalColMap()
	accounts := []lib.Account{
		{ID: 10, Bank: "First National Bank", Budget: lib.BudgetFON, Abbrev: "FNB"},
	}
	defaults := getBaseDefaults()
	categoriesMap := createCategoriesMap(getMockCategories())
	employeesMap := createEmployeesMap([]lib.Employee{
		{ID: "E10", Lastname: "DOE", Firstname: "JOHN", Active: true},
	})
	providersMap := createProvidersMap([]lib.Provider{
		{ID: "P50", Name: "TechCorp Solutions", City: "Faketown"},
	})
	periodsMap := createPeriodsMap(getMockPeriods())

	// Invalid Date format
	row := []string{
		"2025-01-01", "Test", "10", "Office Supplies", "FON", "", "", "card",
		"depenses", "", "", "", "First National Bank",
	}

	_, err := createEntryFromRow(row, colMap, defaults, 1, accounts,
		categoriesMap, employeesMap, providersMap, periodsMap)

	if err == nil || !strings.Contains(err.Error(), "failed to parse date") {
		t.Errorf("Expected date parsing error, got: %v", err)
	}
}

func TestCreateEntryFromRow_MultipleErrors(t *testing.T) {
	colMap := getMinimalColMap()
	accounts := []lib.Account{
		{ID: 10, Bank: "First National Bank", Budget: lib.BudgetFON, Abbrev: "FNB"},
	}
	defaults := getBaseDefaults()
	categoriesMap := createCategoriesMap(getMockCategories())
	employeesMap := createEmployeesMap([]lib.Employee{
		{ID: "E10", Lastname: "DOE", Firstname: "JOHN", Active: true},
	})
	providersMap := createProvidersMap([]lib.Provider{
		{ID: "P50", Name: "TechCorp Solutions", City: "Faketown"},
	})
	periodsMap := createPeriodsMap(getMockPeriods())

	// Row with three errors:
	// 1. Invalid Date: "2025-01-01" (needs "01/01/2025" for `lib.DateLayout`)
	// 2. Both Employee and Provider set (mutual exclusion violation).
	// 3. Invalid Budget: "INVALID_BUDGET"
	row := []string{
		"2025-01-01",         // DATE (Error 1)
		"Test",               // NAME
		"10",                 // AMOUNT
		"Office Supplies",    // CATEGORY
		"INVALID_BUDGET",     // BUDGET (Error 3)
		"John Doe",           // EMPLOYEE (Part of Error 2)
		"TechCorp Solutions", // PROVIDER (Part of Error 2)
		"card",               // PAYMENT
		"depenses",           // KIND
		"", "", "",           // COMMENT, STOCK, PERIOD
		"First National Bank", // BANK
	}

	_, err := createEntryFromRow(row, colMap, defaults, 1, accounts,
		categoriesMap, employeesMap, providersMap, periodsMap)

	if err == nil {
		t.Fatalf("Expected multiple errors, got nil")
	}

	errorString := err.Error()

	// Check for the error from Date parsing
	if !strings.Contains(errorString, "failed to parse date '2025-01-01'") {
		t.Errorf("Expected date parsing error not found in multi-error: %s", errorString)
	}

	// Check for the mutual exclusion error
	if !strings.Contains(errorString, "has both employee ('John Doe') and provider ('TechCorp Solutions') specified") {
		t.Errorf("Expected mutual exclusion error not found in multi-error: %s", errorString)
	}

	// Check for the invalid budget error
	if !strings.Contains(errorString, "invalid budget 'INVALID_BUDGET'") {
		t.Errorf("Expected invalid budget error not found in multi-error: %s", errorString)
	}

	// Check the number of errors
	if strings.Count(errorString, "\n") < 2 { // errors.Join separates the first error from the rest with one newline, and subsequent errors with newlines.
		t.Errorf("Expected at least 3 errors separated by newlines, but found only %d newlines: %s", strings.Count(errorString, "\n"), errorString)
	}
}

func TestGetAccountFromBankBudget_Success(t *testing.T) {
	tests := []struct {
		name     string
		accounts []lib.Account
		bank     string
		budget   lib.Budget
		wantID   int
		wantErr  bool
	}{
		{
			name: "Exact Match FON",
			accounts: []lib.Account{
				{ID: 10, Bank: "First National Bank", Budget: lib.BudgetFON, Abbrev: "FNB"},
			},
			bank:    "First National Bank",
			budget:  lib.BudgetFON,
			wantID:  10,
			wantErr: false,
		},
		{
			name: "Exact Match ASC",
			accounts: []lib.Account{
				{ID: 20, Bank: "Global Reserve", Budget: lib.BudgetASC, Abbrev: "GR"},
			},
			bank:    "Global Reserve",
			budget:  lib.BudgetASC,
			wantID:  20,
			wantErr: false,
		},
		{
			name: "Match Undefined Budget Account",
			accounts: []lib.Account{
				{ID: 30, Bank: "Community Credit Union", Budget: lib.BudgetUndefined, Abbrev: "CCU"},
			},
			bank:    "Community Credit Union",
			budget:  lib.BudgetFON,
			wantID:  30,
			wantErr: false,
		},
		{
			name: "Bank Not Provided (Single Bank Type)",
			accounts: []lib.Account{
				{ID: 20, Bank: "Global Reserve", Budget: lib.BudgetASC, Abbrev: "GR"},
			},
			bank:    "",
			budget:  lib.BudgetASC,
			wantID:  20,
			wantErr: false,
		},
		{
			name: "Failure - Bank Not Provided (Multiple Bank Types)",
			accounts: []lib.Account{
				{ID: 10, Bank: "First National Bank", Budget: lib.BudgetFON, Abbrev: "FNB"},
				{ID: 20, Bank: "Global Reserve", Budget: lib.BudgetASC, Abbrev: "GR"},
			},
			bank:    "",
			budget:  lib.BudgetFON,
			wantID:  0,
			wantErr: true,
		},
		{
			name: "Failure - Ambiguous Account Match",
			// On happy-compta we can have two accounts with different names at the same bank and same budget.
			accounts: []lib.Account{
				{ID: 10, Bank: "Global Reserve", Budget: lib.BudgetASC, Abbrev: "ASC"},
				{ID: 20, Bank: "Global Reserve", Budget: lib.BudgetASC, Abbrev: "ASC"},
			},
			bank:    "Global Reserve",
			budget:  lib.BudgetFON,
			wantID:  0,
			wantErr: true,
		},
		{
			name: "Failure - No Match",
			accounts: []lib.Account{
				{ID: 10, Bank: "First National Bank", Budget: lib.BudgetFON, Abbrev: "FNB"},
			},
			bank:    "Imaginary Bank",
			budget:  lib.BudgetFON,
			wantID:  0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account, err := getAccountFromBankBudget(tt.accounts, tt.bank, tt.budget)

			if (err != nil) != tt.wantErr {
				t.Errorf("getAccountFromBankBudget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && account.ID != tt.wantID {
				t.Errorf("getAccountFromBankBudget() got ID = %d, want %d", account.ID, tt.wantID)
			}
		})
	}
}

func TestParseCSV_Success(t *testing.T) {
	accounts := []lib.Account{
		{ID: 10, Bank: "First National Bank", Budget: lib.BudgetFON, Abbrev: "FNB"},
		{ID: 20, Bank: "Global Reserve", Budget: lib.BudgetASC, Abbrev: "GR"},
	}
	categories := getMockCategories()
	employees := []lib.Employee{
		{ID: "E10", Lastname: "DOE", Firstname: "JOHN", Active: true},
	}
	providers := []lib.Provider{
		{ID: "P50", Name: "TechCorp Solutions", City: "Faketown"},
	}
	periods := getMockPeriods()
	defaults := getBaseDefaults()

	// Mock CSV data with header
	csvData := `
DATE,NAME,AMOUNT,CATEGORY,BUDGET,PROVIDER,BANK,KIND
01/01/2025,Office Supplies Tx,100.50,Office Supplies,FON,TechCorp Solutions,First National Bank,depenses
02/01/2025,Gift Card Purchase,20,Gifts,ASC,,Global Reserve,depenses
`
	r := csv.NewReader(strings.NewReader(csvData))
	r.Comma = ','
	r.Comment = 0

	columnsCfg := CSVColumns{
		Date:     "DATE",
		Name:     "NAME",
		Amount:   "AMOUNT",
		Category: "CATEGORY",
		Budget:   "BUDGET",
		Provider: "PROVIDER",
		Bank:     "BANK",
		Kind:     "KIND",
	}

	// Expected entries (simplified check)
	expectedName1 := "Office Supplies Tx"
	expectedAmount2 := 20.00

	entries, err := parseCSV(r, columnsCfg, defaults, accounts,
		categories, employees, providers, periods)

	if err != nil {
		t.Fatalf("parseCSV failed unexpectedly: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(entries))
	}

	// Check first entry
	if entries[0].Name != expectedName1 {
		t.Errorf("Entry 1 Name mismatch. Got: %s, Want: %s", entries[0].Name, expectedName1)
	}
	if entries[0].Account.ID != 10 { // First National Bank, FON
		t.Errorf("Entry 1 Account ID mismatch. Got: %d, Want: %d", entries[0].Account.ID, 10)
	}

	// Check second entry
	if entries[1].Allocation[0].Amount != expectedAmount2 {
		t.Errorf("Entry 2 Amount mismatch. Got: %.2f, Want: %.2f",
			entries[1].Allocation[0].Amount, expectedAmount2)
	}
	if entries[1].Allocation[0].CategoryID != 200 { // Gifts
		t.Errorf("Entry 2 Category ID mismatch. Got: %d, Want: %d",
			entries[1].Allocation[0].CategoryID, 200)
	}
}

func TestParseCSV_ErrorHandling(t *testing.T) {
	accounts := []lib.Account{
		{ID: 10, Bank: "First National Bank", Budget: lib.BudgetFON, Abbrev: "FNB"},
	}
	categories := getMockCategories()
	employees := []lib.Employee{
		{ID: "E10", Lastname: "DOE", Firstname: "JOHN", Active: true},
	}
	providers := []lib.Provider{
		{ID: "P50", Name: "TechCorp Solutions", City: "Faketown"},
	}
	periods := getMockPeriods()
	defaults := getBaseDefaults()

	// Mock CSV data with errors
	csvData := `
DATE,NAME,AMOUNT,CATEGORY,BUDGET,PROVIDER,BANK
01/01/2025,Valid Tx,100,Office Supplies,FON,TechCorp Solutions,First National Bank
INVALID DATE,Error Date,,,,,
`
	r := csv.NewReader(strings.NewReader(csvData))
	r.Comma = ','
	r.Comment = 0

	columnsCfg := CSVColumns{
		Date: "DATE", Name: "NAME", Amount: "AMOUNT", Category: "CATEGORY", Budget: "BUDGET", Provider: "PROVIDER", Bank: "BANK", Kind: "KIND",
	}

	_, err := parseCSV(r, columnsCfg, defaults, accounts,
		categories, employees, providers, periods)

	if err == nil || !strings.Contains(err.Error(), "failed to process entry on row 2") {
		t.Fatalf("Expected processing error on row 2, but got: %v", err)
	}
}
