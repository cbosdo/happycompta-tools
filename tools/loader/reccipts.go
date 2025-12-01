// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cbosdo/happycompta-tools/lib"
)

// maxReceiptFileSize is 2MB
const maxReceiptFileSize = 2 * 1024 * 1024

// checkAndGetFiles reads all files in a directory, checking file count (max 3) and size (max 2MB) constraints.
func checkAndGetFiles(dir string) (receipts []string, err error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		err = fmt.Errorf("failed to read directory %s: %w", dir, err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		var info os.FileInfo
		filePath := filepath.Join(dir, file.Name())
		info, err = os.Stat(filePath)
		if err != nil {
			err = fmt.Errorf("failed to get file info for %s: %w", filePath, err)
			return
		}

		if info.Size() > maxReceiptFileSize {
			err = fmt.Errorf(
				"receipt file %s is too large (%.2fMB > 2MB)",
				filePath, float64(info.Size())/float64(maxReceiptFileSize),
			)
			return
		}

		receipts = append(receipts, filePath)
	}

	if len(receipts) > 3 {
		return nil, fmt.Errorf("found %d receipt files in %s, but maximum is 3 per entry", len(receipts), dir)
	}

	return
}

// createEmployeeEntryMap creates a map from potential employee full name strings to a list of matching entry indices.
// Employees can be matched lower case using either "Firstname Lastname" or the reverse.
func createEmployeeEntryMap(entries []lib.Entry) map[string][]int {
	employeeMap := make(map[string][]int)
	for i, entry := range entries {
		if emp, ok := entry.Party.(*lib.Employee); ok {
			lnFn := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%s %s", emp.Lastname, emp.Firstname)))
			fnLn := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%s %s", emp.Firstname, emp.Lastname)))

			if lnFn != " " {
				employeeMap[lnFn] = append(employeeMap[lnFn], i)
			}
			if fnLn != " " && lnFn != fnLn {
				employeeMap[fnLn] = append(employeeMap[fnLn], i)
			}
		}
	}
	return employeeMap
}

// addReceipts looks for receipts in the configured folder to attach to the entries.
func addReceipts(receiptsFolder string, entries []lib.Entry) error {
	if receiptsFolder == "" {
		return nil
	}

	items, err := os.ReadDir(receiptsFolder)
	if err != nil {
		return fmt.Errorf("failed to read root receipts folder %s: %w", receiptsFolder, err)
	}

	var subfolders []os.DirEntry
	var rootFiles []os.DirEntry

	for _, item := range items {
		if item.IsDir() {
			subfolders = append(subfolders, item)
		} else {
			rootFiles = append(rootFiles, item)
		}
	}

	// Global Receipts: no nested folder and max three files, add to all entries.
	if len(subfolders) == 0 && len(rootFiles) > 0 {
		allReceipts, err := checkAndGetFiles(receiptsFolder)
		if err != nil {
			return err
		}

		for i := range entries {
			entries[i].Receipts = allReceipts
		}
		return nil
	}

	// Receipts sorted in folders named after one of the entry number (starting from 1) or the employee's full name.
	employeeMap := createEmployeeEntryMap(entries)

	for _, folder := range subfolders {
		folderName := folder.Name()
		folderPath := filepath.Join(receiptsFolder, folderName)

		// Get and validate receipts in the subfolder
		receipts, err := checkAndGetFiles(folderPath)
		if err != nil {
			return fmt.Errorf("error processing receipt folder %s: %w", folderName, err)
		}
		if len(receipts) == 0 {
			continue // Skip empty folders
		}

		applied := false

		// Try if the folder named with entry number.
		if entryNum, err := strconv.Atoi(folderName); err == nil {
			entryIndex := entryNum - 1
			if entryIndex >= 0 && entryIndex < len(entries) {
				entries[entryIndex].Receipts = receipts
				applied = true
			}
		}

		// Folder name matches employee full name.
		if !applied {
			if indices, ok := employeeMap[strings.ToLower(folderName)]; ok {
				for _, index := range indices {
					entries[index].Receipts = receipts
				}
				applied = true
			}
		}
	}

	return nil
}
