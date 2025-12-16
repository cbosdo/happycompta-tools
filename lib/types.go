// SPDX-FileCopyrightText: 2025 SUSE LLC
// SPDX-FileContributor: Cédric Bosdonnat
//
// SPDX-License-Identifier: Apache-2.0

package lib

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Budget int

const (
	BudgetUndefined Budget = iota
	BudgetFON
	BudgetASC
)

func (b Budget) String() string {
	switch b {
	case BudgetFON:
		return "FON"
	case BudgetASC:
		return "ASC"
	}
	return "unknown"
}

func (b *Budget) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*b = BudgetUndefined
		return nil
	}

	var i int
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}

	*b = NewBudget(i)

	if *b == BudgetUndefined && i != 0 && i != 3 {
		return fmt.Errorf("unknown Budget value: %d", i)
	}

	return nil
}

func NewBudget(val int) Budget {
	switch val {
	case 1:
		return BudgetFON
	case 2:
		return BudgetASC
	}
	return BudgetUndefined
}

// NewBudgetFromString creates a Budget value from a string.
// FON and AEP are read as BudgetFON while ASC is read as BudgetASC.
func NewBudgetFromString(s string) Budget {
	upper := strings.ToUpper(s)
	switch upper {
	case "FON":
		fallthrough
	case "AEP":
		return BudgetFON
	case "ASC":
		return BudgetASC
	}
	return BudgetUndefined
}

const (
	KindUndefined Kind = iota
	KindSpend
	KindTake
	KindAllocation
)

type Kind int

func (k Kind) String() string {
	switch k {
	case KindSpend:
		return "depenses"
	case KindTake:
		return "recettes"
	case KindAllocation:
		return "attributions"
	}
	return "unknown"
}

func (k *Kind) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	*k = NewKind(s)

	if *k == KindUndefined && s != "" {
		return fmt.Errorf("unknown Kind value: %s", s)
	}

	return nil
}

func NewKind(s string) Kind {
	switch s {
	case "depenses":
		return KindSpend
	case "recettes":
		return KindTake
	case "attributions":
		return KindAllocation
	}
	return KindUndefined
}

const (
	PeriodStatusUndefined PeriodStatus = iota
	PeriodStatusCurrent
	PeriodStatusProvisionallyClosed
	PeriodStatusDefinitelyClosed
)

type PeriodStatus int

func (s PeriodStatus) String() string {
	switch s {
	case PeriodStatusDefinitelyClosed:
		return "definitely closed"
	case PeriodStatusProvisionallyClosed:
		return "provisionally closed"
	case PeriodStatusCurrent:
		return "current"
	}
	return "unknown"
}

func (s *PeriodStatus) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*s = PeriodStatusUndefined
		return nil
	}

	var i int
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}

	*s = NewPeriodStatus(i)

	if *s == PeriodStatusUndefined && i != 0 {
		return fmt.Errorf("unknown PeriodStatus value: %d", i)
	}

	return nil
}

func NewPeriodStatus(val int) PeriodStatus {
	switch val {
	case 1:
		return PeriodStatusCurrent
	case 2:
		return PeriodStatusProvisionallyClosed
	case 3:
		return PeriodStatusDefinitelyClosed
	}
	return PeriodStatusUndefined
}

type PaymentMethod int

const (
	PaymentMethodUndefined       PaymentMethod = iota
	PaymentMethodCheckReceived   PaymentMethod = 12
	PaymentMethodCash            PaymentMethod = 13
	PaymentMethodCard            PaymentMethod = 14
	PaymentMethodTransfer        PaymentMethod = 15
	PaymentMethodDirectDebit     PaymentMethod = 16
	PaymentMethodCheckEmitted    PaymentMethod = 22
	PaymentMethodCheckAllocation PaymentMethod = 23
)

func (p PaymentMethod) String() string {
	switch p {
	case PaymentMethodCheckReceived:
		return "check received"
	case PaymentMethodCash:
		return "cash"
	case PaymentMethodCard:
		return "card"
	case PaymentMethodTransfer:
		return "transfer"
	case PaymentMethodDirectDebit:
		return "direct debit"
	case PaymentMethodCheckEmitted:
		return "check emitted"
	case PaymentMethodCheckAllocation:
		return "check allocation"
	}
	return "unknown"
}

// NewPaymentMethodFromString converts a string (case-insensitive) into a PaymentMethod value.
func NewPaymentMethodFromString(s string) PaymentMethod {
	lowerS := strings.ToLower(s)

	switch lowerS {
	case "check received":
		return PaymentMethodCheckReceived
	case "cash":
		return PaymentMethodCash
	case "card":
		return PaymentMethodCard
	case "transfer":
		return PaymentMethodTransfer
	case "direct debit":
		return PaymentMethodDirectDebit
	case "check emitted":
		return PaymentMethodCheckEmitted
	case "check allocation":
		return PaymentMethodCheckAllocation
	default:
		return PaymentMethodUndefined
	}
}

// IntBool wraps a boolean and handles 0/1 JSON integers.
type IntBool bool

// UnmarshalJSON implements the json.Unmarshaler interface.
func (b *IntBool) UnmarshalJSON(data []byte) error {
	var intValue int

	if err := json.Unmarshal(data, &intValue); err != nil {
		if string(data) == "null" {
			*b = false
			return nil
		}
		return fmt.Errorf("IntBool field expects an integer (0 or 1), got %s", data)
	}

	// Now, map the integer value to a boolean.
	switch intValue {
	case 1:
		*b = true
	default:
		*b = false
	}
	return nil
}
