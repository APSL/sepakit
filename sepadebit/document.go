package sepadebit

import (
	"crypto/rand"
	"encoding/xml"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"strings"
	"time"
)

// Document is the SEPA format for the document containing all transfers
type Document struct {
	XMLName          xml.Name        `xml:"Document"`
	XMLNs            string          `xml:"xmlns,attr"`
	XMLxsi           string          `xml:"xmlns:xsi,attr"`
	MsgID            string          `xml:"CstmrDrctDbtInitn>GrpHdr>MsgId"`
	CreationDateTime string          `xml:"CstmrDrctDbtInitn>GrpHdr>CreDtTm"`
	TransacNb        int             `xml:"CstmrDrctDbtInitn>GrpHdr>NbOfTxs"`
	CtrlSum          Amount          `xml:"CstmrDrctDbtInitn>GrpHdr>CtrlSum"`
	InitiatingParty  InitiatingParty `xml:"CstmrDrctDbtInitn>GrpHdr>InitgPty"`
	Payments         []*Payment      `xml:"CstmrDrctDbtInitn>PmtInf"`
}

//InitiatingParty is the Initiating Party
type InitiatingParty struct {
	Name   string `xml:"Nm"`
	ID     string `xml:"Id>OrgId>Othr>Id"`
	Scheme string `xml:"Id>OrgId>Othr>SchmeNm>Prtry"`
}

type Creditor struct {
	Name          string        `xml:"Cdtr>Nm"`
	PostalAddress PostalAddress `xml:"Cdtr>PstlAdr,omitempty"`
	IBAN          string        `xml:"CdtrAcct>Id>IBAN"`
	BIC           string        `xml:"CdtrAgt>FinInstnId>BIC"`
	ChargeBearer  string        `xml:"ChrgBr"`
	ID            string        `xml:"CdtrSchmeId>Id>PrvtId>Othr>Id"`
	SchemeName    string        `xml:"CdtrSchmeId>Id>PrvtId>Othr>SchmeNm>Prtry"`
}

type PostalAddress struct {
	Address [2]string `xml:"Adrline,omitempty"`
	Country string    `xml:"Ctry,omitempty"`
}

func NewCreditor() *Creditor {
	c := &Creditor{
		SchemeName: "SEPA",
	}
	return c
}

type Payment struct {
	ID                      string `xml:"PmtInfId"`
	Method                  string `xml:"PmtMtd"`
	TransacNb               int    `xml:"NbOfTxs"`
	CtrlSum                 Amount `xml:"CtrlSum"`
	ServiceLevel            string `xml:"PmtTpInf>SvcLvl>Cd"`
	LocalInstrument         string `xml:"PmtTpInf>LclInstrm>Cd"`
	SequenceType            string `xml:"PmtTpInf>SeqTp"`
	RequestedCollectionDate string `xml:"ReqdColltnDt"`
	*Creditor
	Transactions []Transaction `xml:"DrctDbtTxInf"`
}

func NewPayment() *Payment {
	p := &Payment{
		ServiceLevel:    "SEPA",
		LocalInstrument: "CORE",
		SequenceType:    "RCUR",
	}
	return p
}

type Debtor struct {
	BIC  string `xml:"DbtrAgt>FinInstnId>BIC"`
	Name string `xml:"Dbtr>Nm"`
	IBAN string `xml:"DbtrAcct>Id>IBAN"`
}
type Date time.Time

func (d Date) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	dateString := time.Time(d).Format("2006-01-02")
	e.EncodeElement(dateString, start)
	return nil
}

type Transaction struct {
	ID        string  `xml:"PmtId>EndToEndId"`
	Amount    TAmount `xml:"InstdAmt"`
	MandateID string  `xml:"DrctDbtTx>MndtRltdInf>MndtId"`
	Date      Date    `xml:"DrctDbtTx>MndtRltdInf>DtOfSgntr"`
	Debtor
	RemittanceInfo string `xml:"RmtInf>Ustrd"`
}
type Amount float64

// func (a Amount) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
// 	// dateString := fmt.Sprintf("%2f", a)
// 	dateString := "hola"
// 	e.EncodeElement(dateString, start)
// 	return nil
// }

// func (t TAmount) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
// 	dateString := fmt.Sprintf("%.2f", t.Amount)
// 	e.EncodeElement(dateString, start)
// 	return nil
// }

// func (a Amount) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
// 	// dateString := fmt.Sprintf("%.2f", a)
// 	dateString := "floaaat"
// 	e.EncodeElement(dateString, start)
// 	return nil
// }

// func (t TAmount) MarshalXMLAttr(name xml.Name) (attr xml.Attr, err error) {
// 	attr = xml.Attr{
// 		Name:  name,
// 		Value: "holaa",
// 	}
// 	return
// }

// TAmount is the transaction amount with its currency
type TAmount struct {
	Amount   string `xml:",chardata"`
	Currency string `xml:"Ccy,attr"`
}

func NewDocument() *Document {
	d := &Document{
		XMLNs:  "urn:iso:std:iso:20022:tech:xsd:pain.008.001.02",
		XMLxsi: "http://www.w3.org/2001/XMLSchema-instance",
	}
	t := time.Now()
	d.SetCreationDateTime(t)
	r := make([]byte, 8)
	io.ReadFull(rand.Reader, r)
	d.MsgID = fmt.Sprintf("f-%s-%x", t.Format("20060102"), r)
	return d
}

func (d *Document) SetCreationDateTime(t time.Time) {
	d.CreationDateTime = t.Format("2006-01-02T15:04:05")
}
func (d *Document) SetInitiatingParty(name, id string) {
	d.InitiatingParty.Name = name
	d.InitiatingParty.ID = id
	d.InitiatingParty.Scheme = "SEPA" //fixed
}
func (d *Document) AddPayment(p *Payment) {
	d.Payments = append(d.Payments, p)
}

// func NewCreditor() *Creditor {
// 	c := &Creditor{
// 		SchemeName: "SEPA",
// 	}
// 	return c
// }

// func NewPayment() *Payment {
// 	p := &Payment{
// 		ServiceLevel:    "SEPA",
// 		LocalInstrument: "CORE",
// 		SequenceType:    "RCUR",
// 	}
// 	return p
// }

// func (doc *Document) AddTransaction(id string, amount float64, currency string, creditorName string, creditorIBAN string) error {
// 	if !IsValid(creditorIBAN) {
// 		return errors.New("Invalid creditor IBAN")
// 	}
// 	if DecimalsNumber(amount) > 2 {
// 		return errors.New("Amount 2 decimals only")
// 	}
// 	doc.PaymentTransactions = append(doc.PaymentTransactions, Transaction{
// 		TransacID:           id,
// 		TransacIDe2e:        id,
// 		TransacMotif:        id,
// 		TransacAmount:       TAmount{Amount: amount, Currency: currency},
// 		TransacCreditorName: creditorName,
// 		TransacCreditorIBAN: creditorIBAN,
// 		TransacRegulatory:   "150", // always 150
// 	})
// 	doc.TransacNb++
// 	doc.PaymentInfoTransacNb++

// 	amountCents, e := ToCents(amount)
// 	if e != nil {
// 		return errors.New("In AddTransaction can't convert amount in cents")
// 	}
// 	cumulCents, _ := ToCents(doc.CtrlSum)
// 	if e != nil {
// 		return errors.New("In AddTransaction can't convert control sum in cents")
// 	}

// 	cumulCents += amountCents

// 	cumulEuro, _ := ToEuro(cumulCents)
// 	if e != nil {
// 		return errors.New("In AddTransaction can't convert cumul in euro")
// 	}

// 	doc.CtrlSum = cumulEuro
// 	doc.PaymentInfoCtrlSum = cumulEuro
// 	return nil
// }

// Serialize returns the xml document in byte stream
func (doc *Document) Serialize() ([]byte, error) {
	return xml.Marshal(doc)
}

// PrettySerialize returns the indented xml document in byte stream
func (doc *Document) PrettySerialize() ([]byte, error) {
	return xml.MarshalIndent(doc, "", "  ")
}

// IsValid IBAN
func IsValid(iban string) bool {
	i := new(big.Int)
	t := big.NewInt(10)
	if len(iban) < 4 || len(iban) > 34 {
		return false
	}
	for _, v := range iban[4:] + iban[:4] {
		switch {
		case v >= 'A' && v <= 'Z':
			ch := v - 'A' + 10
			i.Add(i.Mul(i, t), big.NewInt(int64(ch/10)))
			i.Add(i.Mul(i, t), big.NewInt(int64(ch%10)))
		case v >= '0' && v <= '9':
			i.Add(i.Mul(i, t), big.NewInt(int64(v-'0')))
		case v == ' ':
		default:
			return false
		}
	}
	return i.Mod(i, big.NewInt(97)).Int64() == 1
}

// DecimalsNumber returns the number of decimals in a float
func DecimalsNumber(f float64) int {
	s := strconv.FormatFloat(f, 'f', -1, 64)
	p := strings.Split(s, ".")
	if len(p) < 2 {
		return 0
	}
	return len(p[1])
}

// ToCents returns the cents representation in int64
func ToCents(f float64) (int64, error) {
	s := strconv.FormatFloat(f, 'f', 2, 64)
	sc := strings.Replace(s, ".", "", 1)
	return strconv.ParseInt(sc, 10, 64)
}

// ToEuro returns the euro representation in float64
func ToEuro(i int64) (float64, error) {
	d := strconv.FormatInt(i, 10)
	df := d[:len(d)-2] + "." + d[len(d)-2:]
	return strconv.ParseFloat(df, 64)
}
