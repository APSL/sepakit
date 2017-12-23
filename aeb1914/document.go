package aeb1914

import (
	"fmt"
	"time"
)

type InitiatingParty struct {
	ID           string
	Name         string
	CreationDate time.Time
	FileID       string
	Entity       string
	Office       string
}

type Creditor struct {
	ID        string
	Name      string
	AddressD1 string
	AddressD2 string
	AddressD3 string
	Country   string
	Account   string
}

type Debtor struct {
	Entity    string
	Name      string
	AddressD1 string
	AddressD2 string
	AddressD3 string
	Country   string
	IDType    string
	ID        string
	IDTXCode  string
	AccountID string
	Account   string
}
type DebitTransaction struct {
	ID           string
	MandateID    string
	Sequence     string
	CategoryCode string
	Amount       float64
	Date         time.Time
	Debtor       Debtor
	Purpose      string
	Concept      string
}

type DatePayment struct {
	Date               time.Time
	DebitTransactions  []*DebitTransaction
	TotalAmount        float64
	DebitRegisterCount int
	TotalRegisterCount int
}
type CreditorPayments struct {
	Creditor           Creditor
	DatePayments       []*DatePayment
	TotalAmount        float64
	DebitRegisterCount int
	TotalRegisterCount int
}

type Document struct {
	InitiatingParty    *InitiatingParty
	CreditorPayments   []*CreditorPayments
	TotalAmount        float64
	DebitRegisterCount int
	TotalRegisterCount int
}

func NewDocument() *Document {
	return &Document{}
}

func (doc *Document) String() string {
	return fmt.Sprintf("Document Presenter: %s Totals: amount=%f, debits=%d, registers=%d", doc.InitiatingParty.Name, doc.TotalAmount, doc.DebitRegisterCount, doc.TotalRegisterCount)
}
func (dp *DatePayment) String() string {
	return fmt.Sprintf("Payment - date: %s, TotalAmount: %f, TotalDebits: %d", dp.Date, dp.TotalAmount, dp.DebitRegisterCount)
}
func (d *Debtor) String() string {
	return fmt.Sprintf("%s(%s)", d.Name, d.ID)
}
func (t *DebitTransaction) String() string {
	return fmt.Sprintf("Debit Amount: %.2f, Date: %s, Debtor: %s, Concept: %s", t.Amount, t.Date, t.Debtor.Name, t.Concept)
}
