// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"strings"
	"testing"
	"time"
)

func TestExtractIDFromActionsCell(t *testing.T) {
	htmlStr := `<td>
		<a data-id="12345">Edit</a>
		<span>Other Content</span>
		<button data-id="67890">Delete</button>
	</td>`
	td := parseSnippet(t, htmlStr)

	expected := "12345" // Should find the first one encountered
	result := extractIDFromActionsCell(td)

	if result != expected {
		t.Errorf("extractIDFromActionsCell failed. Got: '%s', Expected: '%s'", result, expected)
	}

	// Test case where ID is missing
	htmlStrNoID := `<td><a>Edit</a><span>Other Content</span></td>`
	tdNoID := parseSnippet(t, htmlStrNoID)
	expectedNoID := ""
	resultNoID := extractIDFromActionsCell(tdNoID)

	if resultNoID != expectedNoID {
		t.Errorf("extractIDFromActionsCell (No ID) failed. Got: '%s', Expected: '%s'", resultNoID, expectedNoID)
	}
}

func TestExtractStatusFromStatusCell(t *testing.T) {
	// Status code 1
	htmlStr1 := `<td><span class='label'><span class='hidden'>" . 1 . "</span>En cours</span></td>`
	td1 := parseSnippet(t, htmlStr1)
	expected1 := PeriodStatusCurrent
	result1, err1 := extractStatusFromStatusCell(td1)
	if err1 != nil {
		t.Fatalf("extractStatusFromStatusCell failed for status 1: %v", err1)
	}
	if result1 != expected1 {
		t.Errorf("extractStatusFromStatusCell failed. Got: %d, Expected: %d", result1, expected1)
	}

	// Status code 3
	htmlStr3 := `<td><span class='label label-danger'><span class='hidden'>" . 3 . "</span>Clôture définitive</span></td>`
	td3 := parseSnippet(t, htmlStr3)
	expected3 := PeriodStatusDefinitelyClosed
	result3, err3 := extractStatusFromStatusCell(td3)
	if err3 != nil {
		t.Fatalf("extractStatusFromStatusCell failed for status 3: %v", err3)
	}
	if result3 != expected3 {
		t.Errorf("extractStatusFromStatusCell failed. Got: %d, Expected: %d", result3, expected3)
	}

	// Invalid structure (missing hidden span)
	htmlStrInvalid := `<td><span class='label'>En cours</span></td>`
	tdInvalid := parseSnippet(t, htmlStrInvalid)
	_, errInvalid := extractStatusFromStatusCell(tdInvalid)
	if errInvalid == nil {
		t.Errorf("extractStatusFromStatusCell did not return an error for invalid structure")
	}
}

// =========================================================================
// Main Function Test
// =========================================================================

func TestParsePeriods(t *testing.T) {
	// Re-run with valid data to test successful parsing
	validInputData := `
	<html><body>
	<table id="dt_basic">
    <tbody>
		<div>
            <select class="form-control filter filtre-actif" name="exercice_id">
				<option value="123456">Du 01/01/2025 au 31/12/2025 [En cours]</option>
				<option value="123457">Du 15/06/2023 au 14/06/2024 [Clôture provisoire]</option>
				<option value="123458">Du 01/01/2024 au 31/12/2024 [Clôture définitive]</option>
        </div>
    </tbody>
	</table>
	</body></html>`
	readerValid := strings.NewReader(validInputData)
	periods, err := parsePeriods(readerValid)

	if err != nil {
		t.Fatalf("parsePeriods failed on valid input: %v", err)
	}
	if len(periods) != 3 {
		t.Fatalf("Expected 3 periods, got %d", len(periods))
	}

	// Expected values for Period 1
	expectedP1 := Period{
		ID:     "123456",
		Status: PeriodStatusCurrent,
		Start:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		End:    time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
	}

	if periods[0].ID != expectedP1.ID || periods[0].Status != expectedP1.Status || !periods[0].Start.Equal(expectedP1.Start) || !periods[0].End.Equal(expectedP1.End) {
		t.Errorf("Period 1 mismatch. Got %+v, Expected %+v", periods[0], expectedP1)
	}

	// Expected values for Period 2
	expectedP2 := Period{
		ID:     "123457",
		Status: PeriodStatusProvisionallyClosed,
		Start:  time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC),
		End:    time.Date(2024, 6, 14, 0, 0, 0, 0, time.UTC),
	}

	if periods[1].ID != expectedP2.ID || periods[1].Status != expectedP2.Status || !periods[1].Start.Equal(expectedP2.Start) || !periods[1].End.Equal(expectedP2.End) {
		t.Errorf("Period 2 mismatch. Got %+v, Expected %+v", periods[1], expectedP2)
	}

	// Expected values for Period 3
	expectedP3 := Period{
		ID:     "123458",
		Status: PeriodStatusDefinitelyClosed,
		Start:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		End:    time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
	}

	if periods[2].ID != expectedP3.ID || periods[2].Status != expectedP3.Status || !periods[2].Start.Equal(expectedP3.Start) || !periods[2].End.Equal(expectedP3.End) {
		t.Errorf("Period 3 (Closed, No ID) mismatch. Got %+v, Expected %+v", periods[2], expectedP3)
	}
}
