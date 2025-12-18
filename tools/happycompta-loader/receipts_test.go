// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/cbosdo/happycompta-tools/lib"
)

// Helper function to create a temporary directory and ensure it's cleaned up.
func setupTestDir(t *testing.T, name string) (string, func()) {
	dir, err := os.MkdirTemp("", name)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return dir, func() {
		_ = os.RemoveAll(dir)
	}
}

// Helper function to create a file of a specific size with dummy content.
func createTestFile(t *testing.T, dir, filename string, size int64) string {
	filePath := filepath.Join(dir, filename)
	data := make([]byte, size)
	// Write only a small header for quick file creation
	if size > 0 {
		data[0] = 'T'
		data[size-1] = 'T'
	}
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		t.Fatalf("Failed to create test file %s: %v", filePath, err)
	}
	return filePath
}

func TestCheckAndGetFiles(t *testing.T) {
	const maxTestFileSize = maxReceiptFileSize

	tests := []struct {
		name         string
		fileSizes    []int64 // Sizes of files to create
		expectedFile int     // Expected number of files
		wantErr      bool
		errMsg       string // Substring to check in the error message
	}{
		{
			name:         "Success_SingleFile",
			fileSizes:    []int64{100},
			expectedFile: 1,
			wantErr:      false,
		},
		{
			name:         "Success_MaxFiles",
			fileSizes:    []int64{100, 200, 300},
			expectedFile: 3,
			wantErr:      false,
		},
		{
			name:         "Success_MaxFileSize",
			fileSizes:    []int64{maxTestFileSize},
			expectedFile: 1,
			wantErr:      false,
		},
		{
			name:      "Failure_TooManyFiles",
			fileSizes: []int64{1, 1, 1, 1},
			wantErr:   true,
			errMsg:    "found 4 receipt files",
		},
		{
			name:      "Failure_FileTooLarge",
			fileSizes: []int64{maxTestFileSize + 1},
			wantErr:   true,
			errMsg:    "is too large",
		},
		{
			name:         "Success_EmptyDirectory",
			fileSizes:    []int64{},
			expectedFile: 0,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, cleanup := setupTestDir(t, "checkfiles")
			defer cleanup()

			var expectedReceipts []string
			for i, size := range tt.fileSizes {
				filename := fmt.Sprintf("file_%d.txt", i+1)
				// Create a file, only add to expected if not expecting size error
				filePath := createTestFile(t, dir, filename, size)

				if size <= maxTestFileSize {
					expectedReceipts = append(expectedReceipts, filePath)
				}
			}

			// Add a directory to ensure it's ignored
			if err := os.Mkdir(filepath.Join(dir, "ignored_dir"), 0755); err != nil {
				t.Fatal(err)
			}

			receipts, err := checkAndGetFiles(dir)

			if (err != nil) != tt.wantErr {
				t.Errorf("checkAndGetFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("checkAndGetFiles() got error message = %v, want to contain %s", err, tt.errMsg)
			}

			if !tt.wantErr {
				if len(receipts) != tt.expectedFile {
					t.Errorf("checkAndGetFiles() got %d receipts, want %d", len(receipts), tt.expectedFile)
				}
				if !reflect.DeepEqual(receipts, expectedReceipts) {
					t.Errorf("checkAndGetFiles() got receipts = %v, want %v", receipts, expectedReceipts)
				}
			}
		})
	}
}

func TestCreateEmployeeEntryMap(t *testing.T) {
	// Mock Employee objects for use in entries
	employee1 := lib.Employee{ID: "E1", Lastname: "Doe", Firstname: "John", Active: true}
	employee2 := lib.Employee{ID: "E2", Lastname: "Smith", Firstname: "Alice", Active: true}
	employee3 := lib.Employee{ID: "E3", Lastname: "Jane", Firstname: "Mary", Active: true}

	// Mock Provider object (should be ignored by the map creator)
	provider := lib.Provider{ID: "P1", Name: "Vendor"}

	entries := []lib.Entry{
		// 0: John Doe
		{Party: &employee1},
		// 1: Alice Smith (Different Employee)
		{Party: &employee2},
		// 2: John Doe (Same Employee, different entry)
		{Party: &employee1},
		// 3: Vendor (Provider, ignored)
		{Party: &provider},
		// 4: Empty Party (ignored)
		{},
		// 5: Mary Jane (Case/Order test)
		{Party: &employee3},
	}

	want := map[string][]int{
		"doe john":    {0, 2}, // Lastname Firstname
		"john doe":    {0, 2}, // Firstname Lastname
		"smith alice": {1},
		"alice smith": {1},
		"jane mary":   {5},
		"mary jane":   {5},
	}

	got := createEmployeeEntryMap(entries)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("createEmployeeEntryMap() got = %v, want %v", got, want)
	}

	// Test case sensitivity (should be case-insensitive, map keys are lowercase)
	if _, ok := got["JOHN DOE"]; ok {
		t.Errorf("Expected map key to be lowercase, found 'JOHN DOE'")
	}
}

// Helper to create common mock entries for AddReceipts tests.
func createMockEntries() []lib.Entry {
	return []lib.Entry{
		// Index 0: Alice Smith
		{Name: "Entry 1", Party: &lib.Employee{Lastname: "Smith", Firstname: "Alice"}},
		// Index 1: Alice Smith (Multiple match test)
		{Name: "Entry 2", Party: &lib.Employee{Lastname: "Smith", Firstname: "Alice"}},
		// Index 2: John Doe
		{Name: "Entry 3", Party: &lib.Employee{Lastname: "Doe", Firstname: "John"}},
	}
}

func TestAddReceipts_EmptyFolder(t *testing.T) {
	entries := createMockEntries()
	// Ensure entries start clean
	for i := range entries {
		entries[i].Receipts = nil
	}

	if err := addReceipts("", entries); err != nil {
		t.Errorf("addReceipts with empty folder path failed: %v", err)
	}

	// Assert no receipts were added
	for i, entry := range entries {
		if len(entry.Receipts) != 0 {
			t.Errorf("Entry %d: Expected 0 receipts, got %d", i, len(entry.Receipts))
		}
	}
}

func TestAddReceipts_GlobalMode(t *testing.T) {
	entries := createMockEntries()
	globalOnlyDir, cleanupGlobalOnly := setupTestDir(t, "globalonly")
	defer cleanupGlobalOnly()

	// Setup files directly in the root folder (triggering global mode)
	receipt1 := createTestFile(t, globalOnlyDir, "g1.pdf", 100)
	receipt2 := createTestFile(t, globalOnlyDir, "g2.pdf", 100)
	expectedGlobal := []string{receipt1, receipt2}

	err := addReceipts(globalOnlyDir, entries)
	if err != nil {
		t.Fatalf("addReceipts for global mode failed: %v", err)
	}

	// Check all entries received the global receipts
	for i, entry := range entries {
		if !reflect.DeepEqual(entry.Receipts, expectedGlobal) {
			t.Errorf("Entry %d receipts mismatch. Got: %v, Want: %v", i, entry.Receipts, expectedGlobal)
		}
	}
}

func TestAddReceipts_SubfolderMode_Success(t *testing.T) {
	entries := createMockEntries()
	root, cleanup := setupTestDir(t, "subfolderroot")
	defer cleanup()

	// 1. Setup Index-based Receipts (for Entry 3, index 2)
	idxDir := filepath.Join(root, "3") // Folder name "3" matches index 2 + 1
	if err := os.Mkdir(idxDir, 0755); err != nil {
		t.Fatalf("Failed to create dir %s: %v", idxDir, err)
	}
	createTestFile(t, idxDir, "entry3.png", 100)
	idxReceipts, _ := checkAndGetFiles(idxDir)

	// 2. Setup Employee-based Receipts (for Entry 1 & 2)
	employeeDir := filepath.Join(root, "alice smith") // Employee Full Name (lowercase)
	if err := os.Mkdir(employeeDir, 0755); err != nil {
		t.Fatalf("Failed to create dir %s: %v", employeeDir, err)
	}
	createTestFile(t, employeeDir, "alice.jpg", 100)
	employeeReceipts, _ := checkAndGetFiles(employeeDir)

	// Add an empty subfolder to ensure it's skipped
	if err := os.Mkdir(filepath.Join(root, "empty"), 0755); err != nil {
		t.Fatalf("Failed to create dir: %v", err)
	}

	err := addReceipts(root, entries)
	if err != nil {
		t.Fatalf("addReceipts failed unexpectedly: %v", err)
	}

	// Entry 1 (Index 0) - Matched by "alice smith" folder
	if !reflect.DeepEqual(entries[0].Receipts, employeeReceipts) {
		t.Errorf("Entry 1 receipts mismatch. Got: %v, Want: %v", entries[0].Receipts, employeeReceipts)
	}

	// Entry 2 (Index 1) - Matched by "alice smith" folder
	if !reflect.DeepEqual(entries[1].Receipts, employeeReceipts) {
		t.Errorf("Entry 2 receipts mismatch. Got: %v, Want: %v", entries[1].Receipts, employeeReceipts)
	}

	// Entry 3 (Index 2) - Matched by "3" folder
	if !reflect.DeepEqual(entries[2].Receipts, idxReceipts) {
		t.Errorf("Entry 3 receipts mismatch. Got: %v, Want: %v", entries[2].Receipts, idxReceipts)
	}
}

func TestAddReceipts_SubfolderMode_TooManyReceiptsError(t *testing.T) {
	entries := createMockEntries()
	root, cleanup := setupTestDir(t, "errorroot")
	defer cleanup()

	// Setup Invalid Receipt Folder (too many files: 4 > 3)
	invalidDir := filepath.Join(root, "invalid")
	if err := os.Mkdir(invalidDir, 0755); err != nil {
		t.Fatalf("Failed to create dir %s: %v", invalidDir, err)
	}
	createTestFile(t, invalidDir, "1.pdf", 1)
	createTestFile(t, invalidDir, "2.pdf", 1)
	createTestFile(t, invalidDir, "3.pdf", 1)
	createTestFile(t, invalidDir, "4.pdf", 1)

	// Add another valid folder to ensure the error is found during iteration
	validDir := filepath.Join(root, "valid")
	if err := os.Mkdir(validDir, 0755); err != nil {
		t.Fatalf("Failed to create dir %s: %v", validDir, err)
	}
	createTestFile(t, validDir, "doc.pdf", 100)

	err := addReceipts(root, entries)

	if err == nil {
		t.Fatalf("Expected error for too many receipts in subfolder, but got nil")
	}

	expectedErrSubstring := "found 4 receipt files in "
	if !strings.Contains(err.Error(), expectedErrSubstring) {
		t.Errorf("Expected error to contain '%s', got: %v", expectedErrSubstring, err)
	}
}
