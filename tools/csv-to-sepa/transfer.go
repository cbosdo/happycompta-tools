// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"
	"io"
	"text/template"
	"time"
)

func NewTransferInitiation(ID string, initiator *Party) CustomerCreditTransferInitiation {
	now := time.Now()
	return CustomerCreditTransferInitiation{
		ID:            ID,
		Timestamp:     now.Format("2006-01-02T15:04:05.123Z"),
		ExecutionDate: now.Format("2006-01-02"),
		Initiator:     initiator,
	}
}

type CustomerCreditTransferInitiation struct {
	ID            string
	Timestamp     string
	ExecutionDate string
	Initiator     *Party
	Payments      []*Payment
}

func (c *CustomerCreditTransferInitiation) AddPayment(payment *Payment) {
	if payment.Debtor == nil {
		payment.Debtor = c.Initiator
	}
	if payment.ID == "" {
		payment.ID = fmt.Sprintf("%s/%d", c.ID, len(c.Payments)+1)
	}
	c.Payments = append(c.Payments, payment)
}

func (c *CustomerCreditTransferInitiation) SetTimestamp(timestamp time.Time) {
	c.Timestamp = timestamp.Format("2006-01-02T15:04:05.123Z")
}

func (c *CustomerCreditTransferInitiation) SetExecutionDate(date time.Time) {
	c.ExecutionDate = date.Format("2006-01-02")
}

func (c *CustomerCreditTransferInitiation) Count() int {
	count := 0
	for _, payment := range c.Payments {
		count += len(payment.Transactions)
	}
	return count
}

func (c *CustomerCreditTransferInitiation) Sum() float64 {
	var sum float64
	for _, payment := range c.Payments {
		sum += payment.Sum()
	}
	return sum
}

func (c *CustomerCreditTransferInitiation) Write(wr io.Writer) error {
	t := template.Must(template.New("xml").Parse(transferV3))
	return t.Execute(wr, c)
}

type Payment struct {
	ID           string
	Debtor       *Party
	Transactions []*Transaction
}

func (p Payment) Sum() float64 {
	var sum float64
	for _, transaction := range p.Transactions {
		sum += transaction.Amount
	}
	return sum
}

type Party struct {
	Name string
	IBAN string
	BIC  string
}

type Transaction struct {
	EndToEndID string
	Amount     float64
	Creditor   Party
	Purpose    string
	Info       string
}

const transferV3 = `<?xml version="1.0" encoding="utf-8"?>
<Document xmlns="urn:iso:std:iso:20022:tech:xsd:pain.001.001.03"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
    xsi:schemaLocation="urn:iso:std:iso:20022:tech:xsd:pain.001.001.03 pain.001.001.03.xsd">
    <CstmrCdtTrfInitn>
        <GrpHdr>
            <MsgId>{{ .ID }}</MsgId>
            <CreDtTm>{{ .Timestamp }}</CreDtTm>
            <NbOfTxs>{{ .Count }}</NbOfTxs>
            <CtrlSum>{{ .Sum }}</CtrlSum>
            <InitgPty>
                <Nm>{{ .Initiator.Name }}</Nm>
            </InitgPty>
        </GrpHdr>
{{- range .Payments }}
        <PmtInf>
            <PmtInfId>{{ .ID }}</PmtInfId>
            <PmtMtd>TRF</PmtMtd>
            <BtchBookg>false</BtchBookg>
            <NbOfTxs>{{ .Transactions | len }}</NbOfTxs>
            <CtrlSum>{{ .Sum }}</CtrlSum>
            <ReqdExctnDt>{{ $.ExecutionDate }}</ReqdExctnDt>
            <Dbtr>
                <Nm>{{ .Debtor.Name }}</Nm>
            </Dbtr>
            <DbtrAcct>
                <Id>
                    <IBAN>{{ .Debtor.IBAN }}</IBAN>
                </Id>
            </DbtrAcct>
            <DbtrAgt>
                <FinInstnId>
                    <BIC>{{ .Debtor.BIC }}</BIC>
                </FinInstnId>
            </DbtrAgt>
	{{- range .Transactions }}
            <CdtTrfTxInf>
                <PmtId>
                    <EndToEndId>{{ .EndToEndID }}</EndToEndId>
                </PmtId>
                <Amt>
                    <InstdAmt Ccy="EUR">{{ .Amount }}</InstdAmt>
                </Amt>
                <ChrgBr>SLEV</ChrgBr>
                <CdtrAgt>
                    <FinInstnId>
                        <BIC>{{ .Creditor.BIC }}</BIC>
                    </FinInstnId>
                </CdtrAgt>
                <Cdtr>
                    <Nm>{{ .Creditor.Name }}</Nm>
                </Cdtr>
                <CdtrAcct>
                    <Id>
                        <IBAN>{{ .Creditor.IBAN }}</IBAN>
                    </Id>
                </CdtrAcct>
                <Purp>
                    <Cd>{{ .Purpose }}</Cd>
                </Purp>
                <RmtInf>
                    <Ustrd>{{ .Info }}</Ustrd>
                </RmtInf>
            </CdtTrfTxInf>
	{{- end }}
            </PmtInf>
{{- end }}
    </CstmrCdtTrfInitn>
</Document>
`
