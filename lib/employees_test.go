// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
)

// viewMockReader returns an io.Reader containing a mock view.
func viewMockReader(htmlSnippet string) io.Reader {
	type mockContent struct {
		View string `json:"view"`
	}

	content := mockContent{
		View: htmlSnippet,
	}

	// Marshal the struct to get the correctly escaped JSON string
	jsonData, err := json.Marshal(content)
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal mock JSON: %v", err))
	}

	return bytes.NewReader(jsonData)
}

// TestParseEmployeesResponse tests the function with valid mock data
func TestParseEmployeesResponse(t *testing.T) {
	htmlTable := `
	<table id="tableSalaries"><thead><tr><th style="min-width: 50px"></th><th>Actif</th><th>Justificatifs</th>
	<th>Identifiant Interne</th><th>Site</th><th>Nom</th><th>Pr&eacute;nom</th><th>Email</th><th>Date d&#039;entr&eacute;e</th>
	<th>Date de sortie</th><th class="actionx4 text-center"></th></tr></thead>
	<tbody>
		<tr class="height-39">
			<td class="width-50"></td>
			<td class="text-center"><span class="hide">1</span><img src="green_check.png"></td>
			<td class="bold"></td>
			<td>IntID001</td>
			<td>SiteA</td>
			<td>Doe</td>
			<td>John</td>
			<td>john.d@example.com</td>
			<td></td>
			<td></td>
			<td class="hidden-xs actionx4"><div class="btn-container">
				<a class="btn btn-primary btn-rounded" href="https://app.happy-compta.fr/salaries/edit/100001">
				<i class="fa fa-edit"></i>
				</a></div>
			</td>
		</tr>
		<tr class="height-39">
			<td class="width-50"></td>
			<td class="text-center"><span class="hide">0</span><img src="red_cross.png"></td>
			<td class="bold"></td>
			<td>IntID002</td>
			<td>SiteB</td>
			<td>Smith</td>
			<td>Jane</td>
			<td>jane.s@example.com</td>
			<td></td>
			<td>22/04/2025</td>
			<td class="hidden-xs actionx4"><div class="btn-container">
				<a class="btn btn-primary btn-rounded" href="https://app.happy-compta.fr/salaries/edit/100002">
					<i class="fa fa-edit"></i>
				</a></div>
			</td>
		</tr>
		<tr class="height-39">
			<td class="width-50"></td>
			<td class="text-center"><span class="hide">1</span><img src="green_check.png"></td>
			<td class="bold"></td>
			<td>IntID003</td>
			<td>SiteC</td>
			<td>M&eacute;r&eacute;ncy</td>
			<td>P&eacute;n&eacute;lope</td>
			<td>penelope.m@example.com</td>
			<td></td>
			<td></td>
			<td class="hidden-xs actionx4"><div class="btn-container">
				<a class="btn btn-primary btn-rounded" href="https://app.happy-compta.fr/salaries/edit/100003">
					<i class="fa fa-edit"></i>
				</a></div>
			</td>
		</tr>6
		<tr class="height-39">
			<td class="width-50"></td>
			<td class="text-center"><span class="hide">1</span><img src="green_check.png"></td>
			<td class="bold"></td>
			<td>IntID004</td>
			<td>SiteD</td>
			<td>D&apos;Artagnan</td>
			<td>Fran&ccedil;ois</td>
			<td>francois.d@example.com</td>
			<td></td>
			<td></td>
			<td class="hidden-xs actionx4"><div class="btn-container">
				<a class="btn btn-primary btn-rounded" href="https://app.happy-compta.fr/salaries/edit/100004">
					<i class="fa fa-edit"></i>
				</a></div>
			</td>
		</tr>
	</tbody>
	</table>
	`
	reader := viewMockReader(htmlTable)

	employees, err := parseEmployeesResponse(reader)

	if err != nil {
		t.Fatalf("ParseEmployeesResponse returned an error: %v", err)
	}

	expectedEmployees := []Employee{
		{ID: "100001", Lastname: "Doe", Firstname: "John", Active: true},
		{ID: "100002", Lastname: "Smith", Firstname: "Jane", Active: false},
		{ID: "100003", Lastname: "Méréncy", Firstname: "Pénélope", Active: true},
		{ID: "100004", Lastname: "D'Artagnan", Firstname: "François", Active: true},
	}

	if len(employees) != len(expectedEmployees) {
		t.Fatalf("Expected %d employees, but got %d", len(expectedEmployees), len(employees))
	}

	for i, actual := range employees {
		expected := expectedEmployees[i]
		if actual.ID != expected.ID {
			t.Errorf("Employee %d ID mismatch. Expected: %s, Got: %s", i, expected.ID, actual.ID)
		}
		if actual.Lastname != expected.Lastname {
			t.Errorf("Employee %d Lastname mismatch. Expected: %s, Got: %s", i, expected.Lastname, actual.Lastname)
		}
		if actual.Firstname != expected.Firstname {
			t.Errorf("Employee %d Firstname mismatch. Expected: %s, Got: %s", i, expected.Firstname, actual.Firstname)
		}
		if actual.Active != expected.Active {
			t.Errorf("Employee %d Active status mismatch. Expected: %t, Got: %t", i, expected.Active, actual.Active)
		}
	}
}

// TestParseEmployeesResponse_ErrorHandling tests various error cases
func TestParseEmployeesResponse_ErrorHandling(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedErrorMsg string
	}{
		{
			name:             "Empty Reader",
			input:            "",
			expectedErrorMsg: "failed to decode JSON: EOF",
		},
		{
			name:             "Invalid JSON",
			input:            `{"view": "<html>", invalid_key: 1}`,
			expectedErrorMsg: "failed to decode JSON: invalid character 'i'",
		},
		{
			name:             "Empty View Field",
			input:            `{"view": ""}`,
			expectedErrorMsg: "",
		},
		{
			name:             "Bad HTML (Non-Table Content)",
			input:            `{"view": "<div>Just a div with no data</div>"}`,
			expectedErrorMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader([]byte(tt.input))
			_, err := parseEmployeesResponse(r)

			if tt.expectedErrorMsg == "" && err != nil {
				t.Fatalf("Didn't expect an error, but got %s", err.Error())
			} else if tt.expectedErrorMsg != "" {
				if err == nil {
					t.Fatalf("Expected an error, but got nil")
				}

				// Use strings.Contains to handle subtle differences in error messages across Go versions/platforms
				if !strings.Contains(err.Error(), tt.expectedErrorMsg) {
					t.Errorf("Expected error message containing: %s, but got: %s", tt.expectedErrorMsg, err.Error())
				}
			}
		})
	}
}

// TestParseEmployeesResponse_NoDataInTable tests cases where the table structure is present but empty.
func TestParseEmployeesResponse_NoDataInTable(t *testing.T) {
	noDataView := `
	<table id="tableSalaries"><thead><tr><th style="min-width: 50px"></th><th>Actif</th><th>Justificatifs</th>
	<th>Identifiant Interne</th><th>Site</th><th>Nom</th><th>Pr&eacute;nom</th><th>Email</th>
	<th>Date d&#039;entr&eacute;e</th><th>Date de sortie</th><th class="actionx4 text-center"></th></tr></thead>
	<tbody>
	</tbody>
	</table>
	`
	r := viewMockReader(noDataView)

	employees, err := parseEmployeesResponse(r)

	if len(employees) != 0 {
		t.Errorf("Expected 0 employees, got %d", len(employees))
	}

	if err != nil {
		t.Errorf("No error expected if no data is provided, got %s", err.Error())
	}
}
