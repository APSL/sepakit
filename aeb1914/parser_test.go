package aeb1914

import (
	"fmt"
	"os"
	"testing"

	"golang.org/x/text/encoding/charmap"
)

func TestInitiator(t *testing.T) {
	f, err := os.Open("c19.txt")
	r := charmap.ISO8859_1.NewDecoder().Reader(f)
	if err != nil {
		t.Error("Error opening c19.txt test file")
	}
	parser := NewParser()
	doc, err := parser.Parse(r)
	if err != nil {
		t.Error(err)
		fmt.Println(err)
		t.FailNow()
	}
	if doc.InitiatingParty == nil {
		t.Error("Nil Initiator")
		t.FailNow()
	}
	fmt.Printf("%+v\n", doc.InitiatingParty)
	for _, creditor := range doc.CreditorPayments {
		fmt.Printf("-> Creditor %s\n", creditor.Creditor.ID)
		for _, payment := range creditor.DatePayments {
			fmt.Printf("--> Date Payments %s\n", payment.Date)
			for j, t := range payment.DebitTransactions {
				fmt.Printf("--> Tx%d: %s\n", j, t)
			}
			fmt.Printf("Payment %s\n", payment)
		}
		fmt.Printf("Creditor %+v\n", creditor)
	}
	fmt.Println(doc)
	// fmt.Printf("Totals: amount=%f, debits=%d, registers=%d\n", doc.TotalAmount, doc.DebitRegisterCount, doc.TotalRegisterCount)

}
