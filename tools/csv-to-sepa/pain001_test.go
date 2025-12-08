// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// setupIntegrationTest creates the necessary temporary files and returns their paths.
func setupIntegrationTest(
	t *testing.T, csvContent, outputFileName string,
) (csvPath string, outPath string, cleanup func()) {
	tempDir, err := os.MkdirTemp("", "sepa_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	csvPath = filepath.Join(tempDir, "data.csv")
	outPath = filepath.Join(tempDir, outputFileName)

	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatalf("failed to write CSV: %v", err)
	}

	cleanup = func() {
		_ = os.RemoveAll(tempDir)
	}

	return csvPath, outPath, cleanup
}

// sanitizeXML removes dynamic timestamps and formatting for stable comparison.
func sanitizeXML(xmlContent string) string {
	// Remove CreDtTm (Creation Date/Time)
	xmlContent = regexp.MustCompile(`<CreDtTm>.*?</CreDtTm>`).ReplaceAllString(xmlContent, `<CreDtTm>TIMESTAMP</CreDtTm>`)

	// Remove ReqdExctnDt (Execution Date) which is dynamic (usually today's date)
	xmlContent = regexp.MustCompile(`<ReqdExctnDt>.*?</ReqdExctnDt>`).ReplaceAllString(xmlContent, `<ReqdExctnDt>{{ ExecutionDate }}</ReqdExctnDt>`)

	// Remove all non-essential whitespace for reliable comparison
	xmlContent = strings.ReplaceAll(xmlContent, " ", "")
	xmlContent = strings.ReplaceAll(xmlContent, "\n", "")
	xmlContent = strings.ReplaceAll(xmlContent, "\r", "")
	xmlContent = strings.ReplaceAll(xmlContent, "\t", "")

	return xmlContent
}

func TestIntegration_SimpleTransfer(t *testing.T) {
	csvInput := `id,creditor,iban,bic,amount,info
"payment xxx",John Doe,FR5120041010051631529138143,DPYCNL539SF,123.45,"payment for xxx"
"payment yyy",Joe Tester,FR69 2004 1010 0569 2744 6332 670,KGJW GIOYXXX,12345.67,"payment for yyy"`

	expectedXML := `<?xml version="1.0" encoding="utf-8"?>
<Document xmlns="urn:iso:std:iso:20022:tech:xsd:pain.001.001.03"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
    xsi:schemaLocation="urn:iso:std:iso:20022:tech:xsd:pain.001.001.03 pain.001.001.03.xsd">
    <CstmrCdtTrfInitn>
        <GrpHdr>
            <MsgId>batch/1</MsgId>
            <CreDtTm>TIMESTAMP</CreDtTm>
            <NbOfTxs>2</NbOfTxs>
            <CtrlSum>12469.12</CtrlSum>
            <InitgPty>
                <Nm>Issuer</Nm>
            </InitgPty>
        </GrpHdr>
        <PmtInf>
            <PmtInfId>batch/1/1</PmtInfId>
            <PmtMtd>TRF</PmtMtd>
            <BtchBookg>false</BtchBookg>
            <NbOfTxs>2</NbOfTxs>
            <CtrlSum>12469.12</CtrlSum>
            <ReqdExctnDt>{{ ExecutionDate }}</ReqdExctnDt>
            <Dbtr>
                <Nm>Issuer</Nm>
            </Dbtr>
            <DbtrAcct>
                <Id>
                    <IBAN>FR7420041010058652109911007</IBAN>
                </Id>
            </DbtrAcct>
            <DbtrAgt>
                <FinInstnId>
                    <BIC>PMXNCXV94RH</BIC>
                </FinInstnId>
            </DbtrAgt>
            <CdtTrfTxInf>
                <PmtId>
                    <EndToEndId>payment xxx</EndToEndId>
                </PmtId>
                <Amt>
                    <InstdAmt Ccy="EUR">123.45</InstdAmt>
                </Amt>
                <ChrgBr>SLEV</ChrgBr>
                <CdtrAgt>
                    <FinInstnId>
                        <BIC>DPYCNL539SF</BIC>
                    </FinInstnId>
                </CdtrAgt>
                <Cdtr>
                    <Nm>John Doe</Nm>
                </Cdtr>
                <CdtrAcct>
                    <Id>
                        <IBAN>FR5120041010051631529138143</IBAN>
                    </Id>
                </CdtrAcct>
                <Purp>
                    <Cd>REFU</Cd>
                </Purp>
                <RmtInf>
                    <Ustrd>payment for xxx</Ustrd>
                </RmtInf>
            </CdtTrfTxInf>
            <CdtTrfTxInf>
                <PmtId>
                    <EndToEndId>payment yyy</EndToEndId>
                </PmtId>
                <Amt>
                    <InstdAmt Ccy="EUR">12345.67</InstdAmt>
                </Amt>
                <ChrgBr>SLEV</ChrgBr>
                <CdtrAgt>
                    <FinInstnId>
                        <BIC>KGJWGIOYXXX</BIC>
                    </FinInstnId>
                </CdtrAgt>
                <Cdtr>
                    <Nm>Joe Tester</Nm>
                </Cdtr>
                <CdtrAcct>
                    <Id>
                        <IBAN>FR6920041010056927446332670</IBAN>
                    </Id>
                </CdtrAcct>
                <Purp>
                    <Cd>REFU</Cd>
                </Purp>
                <RmtInf>
                    <Ustrd>payment for yyy</Ustrd>
                </RmtInf>
            </CdtTrfTxInf>
            </PmtInf>
    </CstmrCdtTrfInitn>
</Document>`

	// Parameters parsed into Config struct
	cfg := Config{
		BatchID: "batch/1",
		Debtor: Party{
			Name: "Issuer",
			IBAN: "FR7420041010058652109911007",
			BIC:  "PMXNCXV94RH",
		},
		CSV: CsvConfig{
			Columns: ColumnsConfig{
				Creditor:   "creditor",
				IBAN:       "iban",
				BIC:        "bic",
				EndToEndID: "id",
				Amount:     "amount",
				Info:       "info",
			},
		},
	}

	csvPath, outPath, cleanup := setupIntegrationTest(t, csvInput, "output.xml")
	defer cleanup()

	// Set the correct output path in the configuration
	cfg.Output = outPath

	// Execute the tool's core logic
	if err := toPain001(cfg, csvPath); err != nil {
		t.Fatalf("toPain001 failed: %v", err)
	}

	// Read the generated output
	generatedData, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read generated output: %v", err)
	}

	sanitizedGenerated := sanitizeXML(string(generatedData))
	sanitizedExpected := sanitizeXML(string(expectedXML))

	if sanitizedGenerated != sanitizedExpected {
		t.Errorf("Generated XML does not match expected XML.")
		t.Logf("--- Expected (Sanitized) ---\n%s", sanitizedExpected)
		t.Logf("--- Got (Sanitized) ---\n%s", sanitizedGenerated)
	}
}
