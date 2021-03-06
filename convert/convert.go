package convert

import (
	"fmt"
	"io"
	"unicode"

	"github.com/apsl/sepakit/aeb1914"
	"github.com/apsl/sepakit/sepadebit"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

//Latin1DebitTxtToXML creates XML SEPA Document from TXT Document
//Transforms input and output to ISO-8859-1
func Latin1DebitTxtToXML(in io.Reader, out io.Writer) error {
	r := charmap.ISO8859_1.NewDecoder().Reader(in)
	parser := aeb1914.NewParser()

	doctxt, err := parser.Parse(r)
	if err != nil {
		return err
	}

	docxml := DebitTxtToXML(doctxt)

	err = docxml.WriteLatin1(out)
	if err != nil {
		return err
	}
	return nil
}

//Latin1DebitTxtToXMLDoc creates XML SEPA Document from TXT Document
//Transforms input to ISO-8859-1
func Latin1DebitTxtToXMLDoc(in io.Reader) (*sepadebit.Document, error) {
	r := charmap.ISO8859_1.NewDecoder().Reader(in)
	parser := aeb1914.NewParser()

	doctxt, err := parser.Parse(r)
	if err != nil {
		return nil, err
	}

	docxml := DebitTxtToXML(doctxt)
	return docxml, nil
}

//DebitTxtToXML creates XML SEPA Document from TXT Document
func DebitTxtToXML(doctxt *aeb1914.Document) *sepadebit.Document {

	docxml := sepadebit.NewDocument()
	docxml.SetInitiatingParty(doctxt.InitiatingParty.Name, doctxt.InitiatingParty.ID)

	// conversion fist version. Should be moved to sepadebit package
	for _, cp := range doctxt.CreditorPayments {
		for _, dp := range cp.DatePayments {
			c := sepadebit.Creditor{
				ID:   cp.Creditor.ID,
				Name: cp.Creditor.Name,
				IBAN: cp.Creditor.Account,
				PostalAddress: sepadebit.PostalAddress{
					Country: cp.Creditor.Country,
					Address: [2]string{cp.Creditor.AddressD1, cp.Creditor.AddressD2},
				},
				SchemeName:   "SEPA",
				BIC:          "CAIXESBBXXX", // hardcoded, we need a BIC library
				ChargeBearer: "SLEV",
			}
			p := sepadebit.Payment{
				Creditor:                &c,
				CtrlSum:                 fmt.Sprintf("%.2f", cp.TotalAmount),
				TransacNb:               cp.DebitRegisterCount,
				RequestedCollectionDate: dp.Date.Format("2006-01-02"),
				ID:                      fmt.Sprintf("rem%s1", dp.Date.Format("20060102")),
				Method:                  "DD",
				ServiceLevel:            "SEPA",
				LocalInstrument:         "CORE",
				SequenceType:            "RCUR",
			}
			for _, dt := range dp.DebitTransactions {
				t := sepadebit.Transaction{
					ID:        dt.ID,
					MandateID: dt.MandateID,
					Date:      sepadebit.Date(dt.Date),
					Debtor: sepadebit.Debtor{
						IBAN: dt.Debtor.Account,
						BIC:  dt.Debtor.Entity,
						Name: stripSepa(dt.Debtor.Name),
					},
					Amount: sepadebit.TAmount{
						Amount:   fmt.Sprintf("%.2f", dt.Amount),
						Currency: "EUR",
					},
					RemittanceInfo: stripSepa(dt.Concept),
				}
				p.Transactions = append(p.Transactions, t)
			}
			docxml.Payments = append(docxml.Payments, &p)
			docxml.CtrlSum = fmt.Sprintf("%.2f", doctxt.TotalAmount)
			docxml.TransacNb = doctxt.DebitRegisterCount
		}

	}
	return docxml
}

//stripSepa implements a filter to remove not alowed chars
func stripSepa(str string) string {
	isMn := func(r rune) bool {
		return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
	}
	isNotSepa := func(r rune) bool {
		return r > 126
	}
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isMn), transform.RemoveFunc(isNotSepa), norm.NFC)
	str, _, _ = transform.String(t, str)
	return str
}
