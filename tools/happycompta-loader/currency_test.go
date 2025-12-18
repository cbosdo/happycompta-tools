// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"strings"
	"testing"
)

func TestParseAmount(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{
			name:    "Standard US Format",
			input:   "1234.56",
			want:    1234.56,
			wantErr: false,
		},
		{
			name:    "Standard US Format with Comma Thousand Separator",
			input:   "1,234.56",
			want:    1234.56,
			wantErr: false,
		},
		{
			name:    "European Format",
			input:   "1234,56",
			want:    1234.56,
			wantErr: false,
		},
		{
			name:    "European Format with Space Separator and Symbol",
			input:   "1 234,56 €",
			want:    1234.56,
			wantErr: false,
		},
		{
			name:    "European Format with NBSP",
			input:   "1 234,56",
			want:    1234.56,
			wantErr: false,
		},
		{
			name:    "No decimal",
			input:   "1000",
			want:    1000.00,
			wantErr: false,
		},
		{
			name:    "Empty String",
			input:   "",
			want:    0,
			wantErr: true,
		},
		{
			name:    "Invalid Characters Left",
			input:   "100.50abc",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAmount(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseAmount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseAmount() got = %f, want %f", got, tt.want)
			}
			if tt.wantErr && !strings.Contains(err.Error(), "failed to parse amount") && !strings.Contains(err.Error(), "missing or empty") {
				t.Errorf("parseAmount() got unexpected error message: %v", err)
			}
		})
	}
}
