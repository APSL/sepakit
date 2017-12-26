package sepadebit

import (
	"crypto/rand"
	"encoding/xml"
	"fmt"
	"io"
	"time"

	"golang.org/x/text/encoding/charmap"
)

// Document is the SEPA format for the document containing all transfers
type Document struct {
	XMLName          xml.Name        `xml:"Document"`
	XMLNs            string          `xml:"xmlns,attr"`
	XMLxsi           string          `xml:"xmlns:xsi,attr"`
	MsgID            string          `xml:"CstmrDrctDbtInitn>GrpHdr>MsgId"`
	CreationDateTime string          `xml:"CstmrDrctDbtInitn>GrpHdr>CreDtTm"`
	TransacNb        int             `xml:"CstmrDrctDbtInitn>GrpHdr>NbOfTxs"`
	CtrlSum          string          `xml:"CstmrDrctDbtInitn>GrpHdr>CtrlSum"`
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
	CtrlSum                 string `xml:"CtrlSum"`
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

//WriteBytes returns XML Serialized document in byte stream
func (d *Document) WriteBytes() ([]byte, error) {
	return xml.MarshalIndent(d, "", "  ")
}

//WriteLatin1 writes ISO8859-1 XML document to io.Writer argument
func (d *Document) WriteLatin1(w io.Writer) error {

	data, err := d.WriteBytes()
	if err != nil {
		return err
	}
	wl1 := charmap.ISO8859_1.NewEncoder().Writer(w)
	header := []byte(`<?xml version="1.0" encoding="iso-8859-1"?>` + "\n")
	wl1.Write(header)
	wl1.Write(data)
	return nil
}
