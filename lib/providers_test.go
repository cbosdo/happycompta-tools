// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"strings"
	"testing"
)

const mockProvidersHTML = `
<html><body>
<table id="dt_basic" width="100%">
    <thead>
        <tr>
            <th>Nom</th><th>Adresse</th><th>Code postal</th><th>Ville</th>
            <th>Téléphone</th><th>Email</th><th>Commentaire</th><th>Relation</th>
            <th class="noPdf"></th>
        </tr>
    </thead>
    <tbody>
        <!-- Provider 1: Not Archived (Archiver button present) -->
        <tr>
            <td>Software Solutions Inc.</td>
            <td>123 Tech Avenue, Suite 100</td>
            <td>90001</td>
            <td>Los Angeles</td>
            <td>+1 555-123-4567</td>
            <td>info@softsol.com</td>
            <td>
                <span>
                    Primary software vendor. Contract renews Q3.
                </span>
            </td>
            <td></td>
            <td class="hidden-xs actionx4">
                <a data-id="P7730" href="/fournisseurs/edit/7730">Edit</a>
                <a title="Archiver ce fournisseur" href="/fournisseurs/archivage/7730">Archive</a>
                <a title="Supprimer ce fournisseur" href="/fournisseurs/delete/7730">Delete</a>
            </td>
        </tr>
        <!-- Provider 2: Archived (Désarchiver button present) -->
        <tr style="background-color: rgba(255, 153, 51, 0.3) !important;">
            <td>Creative Design Studio</td>
            <td>45 Rue des Arts, Apt 2B</td>
            <td>75003</td>
            <td>Paris</td>
            <td>+33 1 40 50 60 70</td>
            <td>contact@creativedesign.fr</td>
            <td>
                <span>
                    Archived due to low project volume in 2024.
                </span>
            </td>
            <td></td>
            <td class="hidden-xs actionx4">
                <a data-id="P4481" href="/fournisseurs/edit/4481">Edit</a>
                <a title="Désarchiver ce fournisseur" data-archive="1" href="/fournisseurs/desarchivage/4481">Unarchive</a>
                <a title="Supprimer ce fournisseur" href="/fournisseurs/delete/4481">Delete</a>
            </td>
        </tr>
        <!-- Provider 3: Missing fields/ID (Should still parse successfully) -->
        <tr>
            <td>Local Catering</td>
            <td>55 Market Street</td>
            <td>00000</td>
            <td>Unknownville</td>
            <td>(999) 555-1212</td>
            <td>catering@local.com</td>
            <td>
                <span>
                    Used for monthly office luncheons.
                </span>
            </td>
            <td></td>
            <td class="hidden-xs actionx4">
                <!-- No ID or archive button -->
            </td>
        </tr>
        <!-- Invalid row structure (should be skipped by length check) -->
        <tr><td>Only one cell</td></tr>
    </tbody>
</table>
</body></html>`

func TestParseProviders_Success(t *testing.T) {
	reader := strings.NewReader(mockProvidersHTML)
	providers, err := parseProviders(reader)

	if err != nil {
		t.Fatalf("parseProviders failed unexpectedly: %v", err)
	}
	if len(providers) != 3 {
		t.Fatalf("Expected 3 providers, got %d", len(providers))
	}

	// Expected Provider 1 (Not Archived)
	p1 := providers[0]
	if p1.ID != "P7730" {
		t.Errorf("P1 ID mismatch. Got: %s", p1.ID)
	}
	if p1.Name != "Software Solutions Inc." {
		t.Errorf("P1 Name mismatch. Got: %s", p1.Name)
	}
	if p1.ZipCode != "90001" {
		t.Errorf("P1 ZipCode mismatch. Got: %s", p1.ZipCode)
	}
	if p1.Archived != false {
		t.Errorf("P1 Archived status mismatch. Expected: false, Got: %t", p1.Archived)
	}
	if p1.Email != "info@softsol.com" {
		t.Errorf("P1 Email mismatch. Expected: 'info@softsol.com', Got: %s", p1.Email)
	}
	if p1.Comment != "Primary software vendor. Contract renews Q3." {
		t.Errorf("P1 Comment mismatch. Got: %s", p1.Comment)
	}

	// Expected Provider 2 (Archived)
	p2 := providers[1]
	if p2.ID != "P4481" {
		t.Errorf("P2 ID mismatch. Got: %s", p2.ID)
	}
	if p2.Name != "Creative Design Studio" {
		t.Errorf("P2 Name mismatch. Got: %s", p2.Name)
	}
	if p2.City != "Paris" {
		t.Errorf("P2 City mismatch. Got: %s", p2.City)
	}
	if p2.Email != "contact@creativedesign.fr" {
		t.Errorf("P2 Email mismatch. Got: %s", p2.Email)
	}
	if p2.Archived != true {
		t.Errorf("P2 Archived status mismatch. Expected: true, Got: %t", p2.Archived)
	}
	if p2.Comment != "Archived due to low project volume in 2024." {
		t.Errorf("P2 Comment mismatch. Got: %s", p2.Comment)
	}

	// Expected Provider 3 (Missing ID, Not Archived)
	p3 := providers[2]
	if p3.ID != "" {
		t.Errorf("P3 ID mismatch. Expected: empty, Got: %s", p3.ID)
	}
	if p3.Name != "Local Catering" {
		t.Errorf("P3 Name mismatch. Got: %s", p3.Name)
	}
	if p3.ZipCode != "00000" {
		t.Errorf("P3 ZipCode mismatch. Got: %s", p3.ZipCode)
	}
	if p3.Comment != "Used for monthly office luncheons." {
		t.Errorf("P3 Comment mismatch. Got: %s", p3.Comment)
	}
	if p3.Archived != false {
		t.Errorf("P3 Archived status mismatch. Expected: false, Got: %t", p3.Archived)
	}
}

func TestParseProviders_MissingTable(t *testing.T) {
	htmlStr := `<html><body><div id="content">No table data here.</div></body></html>`
	reader := strings.NewReader(htmlStr)
	_, err := parseProviders(reader)

	if err == nil {
		t.Fatal("Expected an error for missing table, but got nil")
	}
	expectedErr := "could not find the table listing the providers"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Expected error containing '%s', got: %v", expectedErr, err)
	}
}

func TestParseProviders_ShortRow(t *testing.T) {
	htmlStr := `
	<html><body><table id="dt_basic"><tbody>
		<tr><td>Cell 1</td><td>Cell 2</td></tr>
	</tbody></table></body></html>`
	reader := strings.NewReader(htmlStr)

	providers, err := parseProviders(reader)

	if err != nil {
		t.Fatalf("Expected no error for short row, but got: %v", err)
	}
	if len(providers) != 0 {
		t.Fatalf("Expected 0 providers (row should be skipped), got %d", len(providers))
	}
}
