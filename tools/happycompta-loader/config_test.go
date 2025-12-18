// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package main

import "testing"

func TestGetSingleRune(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldName string
		wantRune  rune
		wantErr   bool
	}{
		{
			name:      "Valid ASCII Character",
			value:     ",",
			fieldName: "separator",
			wantRune:  ',',
			wantErr:   false,
		},
		{
			name:      "Valid Multi-byte Character (Unicode)",
			value:     "§", // Unicode character
			fieldName: "comment char",
			wantRune:  '§',
			wantErr:   false,
		},
		{
			name:      "Empty String",
			value:     "",
			fieldName: "separator",
			wantRune:  0, // Expected result for empty string
			wantErr:   false,
		},
		{
			name:      "Too Many Characters",
			value:     "||",
			fieldName: "separator",
			wantRune:  0,
			wantErr:   true,
		},
		{
			name:      "Whitespace String",
			value:     " ",
			fieldName: "separator",
			wantRune:  ' ',
			wantErr:   false,
		},
		{
			name:      "Multiple Multi-byte Characters",
			value:     "§§",
			fieldName: "comment char",
			wantRune:  0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRune, err := getSingleRune(tt.value, tt.fieldName)

			if (err != nil) != tt.wantErr {
				t.Errorf("getSingleRune() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRune != tt.wantRune {
				t.Errorf("getSingleRune() gotRune = %c, want %c", gotRune, tt.wantRune)
			}
		})
	}
}
