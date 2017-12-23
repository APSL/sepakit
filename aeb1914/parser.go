package aeb1914

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math"
	"strconv"
	"strings"
	"time"
)

type Parser struct {
	doc             *Document
	currentPayment  *DatePayment
	currentCreditor *CreditorPayments
}

func NewParser() *Parser {
	return &Parser{}
}

//Parse takes a io.Reader with SEPA 19-14 contents in iso-8859 encoding
func (p *Parser) Parse(r io.Reader) (doc *Document, err error) {
	p.doc = NewDocument()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := []rune(scanner.Text())
		regCode := string(line[:2])
		switch regCode {
		case "01":
			err = p.parseInitiatingParty(line)
			if err != nil {
				log.Println(err)
				return
			}
		case "02":
			err = p.parsePaymentHeader(line)
			if err != nil {
				log.Println(err)
				return
			}
		case "03":
			err = p.parseDebitTransaction(line)
			if err != nil {
				log.Println(err)
				return
			}
		case "04":
			err = p.parsePaymentTotals(line)
			if err != nil {
				log.Println(err)
				return
			}
		case "05":
			err = p.parseCreditorTotals(line)
			if err != nil {
				log.Println(err)
				return
			}
		case "99":
			err = p.parseTotals(line)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
	doc = p.doc
	return
}

func (p *Parser) countRegister() {
	p.doc.TotalRegisterCount++
	if p.currentPayment != nil {
		p.currentPayment.TotalRegisterCount++
	}
	if p.currentCreditor != nil {
		p.currentCreditor.TotalRegisterCount++
	}
}

func (p *Parser) countDebitRegister() {
	p.doc.DebitRegisterCount++
	if p.currentPayment != nil {
		p.currentPayment.DebitRegisterCount++
	}
	if p.currentCreditor != nil {
		p.currentCreditor.DebitRegisterCount++
	}
}

func (p *Parser) addDebitAmount(amount float64) {
	p.doc.TotalAmount += amount
	if p.currentPayment != nil {
		p.currentPayment.TotalAmount += amount
	}
	if p.currentCreditor != nil {
		p.currentCreditor.TotalAmount += amount
	}
}

func (p *Parser) parseInitiatingParty(line []rune) (err error) {
	i := &InitiatingParty{}
	dataNum := string(line[7:10])
	if dataNum != "001" {
		return fmt.Errorf("Expected 003 data number but found %s\n", dataNum)
	}
	i.ID = getString(line[10:45])
	i.Name = getString(line[45:115])
	i.FileID = getString(line[123:158])
	i.CreationDate, err = getDate(line[115:123])
	if err != nil {
		log.Printf("Error parsing file creation date: %s\n", err)
	}
	i.Entity = getString(line[158:162])
	i.Office = getString(line[162:166])
	p.doc.InitiatingParty = i
	p.countRegister()
	return
}

//Parses Creditor Date Payment
func (p *Parser) parsePaymentHeader(line []rune) (err error) {
	dataNum := string(line[7:10])
	if dataNum != "002" {
		return fmt.Errorf("Expected 002 datanum for Payment but found %s\n", dataNum)
	}
	cp := p.currentCreditor
	if cp == nil {
		cp = &CreditorPayments{}
	}
	cp.Creditor.ID = getString(line[10:45])
	cp.Creditor.Name = getString(line[53:123])
	date, err := getDate(line[45:53])
	if err != nil {
		log.Printf("Error parsing Payment date: %s\n", err)
	}
	dp := &DatePayment{Date: date}
	cp.DatePayments = append(cp.DatePayments, dp)
	cp.Creditor.AddressD1 = getString(line[123:173])
	cp.Creditor.AddressD2 = getString(line[173:223])
	cp.Creditor.AddressD3 = getString(line[223:263])
	cp.Creditor.Country = getString(line[263:265])
	cp.Creditor.Account = getString(line[265:299])
	p.doc.CreditorPayments = append(p.doc.CreditorPayments, cp)
	p.currentCreditor = cp
	p.currentPayment = dp
	p.countRegister()
	return
}

func (p *Parser) parseDebitTransaction(line []rune) (err error) {
	if p.currentPayment == nil {
		return fmt.Errorf("Parser: got transaction line with no current payment\n")
	}
	t := &DebitTransaction{}
	dataNum := string(line[7:10])
	if dataNum != "003" {
		return fmt.Errorf("Expected 003 datanum for debit transaction but found %s\n", dataNum)
	}
	t.ID = getString(line[10:45])
	t.MandateID = getString(line[45:80])
	t.Sequence = getString(line[80:84])
	t.CategoryCode = getString(line[84:88])
	t.Amount, err = getMoney(line[88:99])
	if err != nil {
		log.Printf("Error parsing Payment Amount: %s\n", err)
	}
	t.Date, err = getDate(line[99:107])
	if err != nil {
		log.Printf("Error parsing Payment date: %s\n", err)
	}
	t.Debtor.Entity = getString(line[107:118])
	t.Debtor.Name = getString(line[118:188])
	t.Debtor.AddressD1 = getString(line[188:238])
	t.Debtor.AddressD1 = getString(line[238:288])
	t.Debtor.AddressD1 = getString(line[288:328])
	t.Debtor.Country = getString(line[328:330])
	t.Debtor.IDType = getString(line[330:331])
	t.Debtor.ID = getString(line[331:367])
	t.Debtor.IDTXCode = getString(line[367:402])
	t.Debtor.AccountID = getString(line[402:403])
	t.Debtor.Account = getString(line[403:437])
	t.Purpose = getString(line[437:441])
	t.Concept = getString(line[441:581])
	p.currentPayment.DebitTransactions = append(p.currentPayment.DebitTransactions, t)
	p.countDebitRegister()
	p.countRegister()
	p.addDebitAmount(t.Amount)
	return
}

func (p *Parser) parsePaymentTotals(line []rune) (err error) {
	if p.currentPayment == nil {
		return fmt.Errorf("Received date payment total line with no current Payment\n")
	}
	if p.currentCreditor == nil {
		return fmt.Errorf("Received date payment total line with no current Creditor\n")
	}
	creditorID := getString(line[02:37])
	date, err := getDate(line[37:45])
	if err != nil {
		return
	}
	totalAmount, err := getMoney(line[45:62])
	if err != nil {
		return
	}
	debitRegisterCount, err := getInt(line[62:70])
	if err != nil {
		return
	}
	totalRegisterCount, err := getInt(line[70:80])
	if err != nil {
		return
	}
	//test parsed values against previous ones
	if creditorID != p.currentCreditor.Creditor.ID {
		return fmt.Errorf("Received totals line with different Creditor ID: %s. Payment.Creditor.ID: %s\n", creditorID, p.currentCreditor.Creditor.ID)
	}
	if date != p.currentPayment.Date {
		return fmt.Errorf("Received totals line with different date: %s. (Payment.date=%s)\n", date, p.currentPayment.Date)
	}
	if math.Abs(totalAmount-p.currentPayment.TotalAmount) > 0.01 {
		return fmt.Errorf("Calculated amount = %f diferent from parsed amount = %f", p.currentPayment.TotalAmount, totalAmount)
	}
	p.currentPayment.TotalAmount = totalAmount //write parsed amount, as calculated is a float intended for comprovation
	//test register count
	if debitRegisterCount != len(p.currentPayment.DebitTransactions) {
		return fmt.Errorf("Debit transactions on totals line = %d. Parsed debit transactions = %d", debitRegisterCount, len(p.currentPayment.DebitTransactions))
	}
	p.countRegister()
	p.currentPayment.DebitRegisterCount = debitRegisterCount
	if p.currentPayment.TotalRegisterCount != totalRegisterCount {
		return fmt.Errorf("Parsed number of register differs: %d - %d", p.currentPayment.TotalRegisterCount, totalRegisterCount)
	}
	p.currentPayment = nil
	return
}

func (p *Parser) parseCreditorTotals(line []rune) (err error) {
	if p.currentCreditor == nil {
		return fmt.Errorf("Received creditor totals line (05) with no current Creditor\n")
	}
	creditorID := getString(line[02:37])
	totalAmount, err := getMoney(line[37:54])
	if err != nil {
		return
	}
	debitRegisterCount, err := getInt(line[54:62])
	if err != nil {
		return
	}
	totalRegisterCount, err := getInt(line[62:72])
	if err != nil {
		return
	}
	//test parsed values against previous ones
	if creditorID != p.currentCreditor.Creditor.ID {
		return fmt.Errorf("Received totals line with different Creditor ID: %s. Payment.Creditor.ID: %s\n", creditorID, p.currentCreditor.Creditor.ID)
	}
	if math.Abs(totalAmount-p.currentCreditor.TotalAmount) > 0.01 {
		return fmt.Errorf("Calculated amount = %f diferent from parsed amount = %f", p.currentCreditor.TotalAmount, totalAmount)
	}
	p.currentCreditor.TotalAmount = totalAmount
	//test register count
	if debitRegisterCount != p.currentCreditor.DebitRegisterCount {
		return fmt.Errorf("Debit transactions on totals line = %d. Parsed debit transactions = %d", debitRegisterCount, p.currentCreditor.DebitRegisterCount)
	}
	p.currentCreditor.DebitRegisterCount = debitRegisterCount
	p.countRegister()
	if p.currentCreditor.TotalRegisterCount != totalRegisterCount {
		return fmt.Errorf("Parsed number of register differs: %d - %d", p.currentCreditor.TotalRegisterCount, totalRegisterCount)
	}
	p.currentCreditor = nil
	return
}

func (p *Parser) parseTotals(line []rune) (err error) {
	totalAmount, err := getMoney(line[02:19])
	if err != nil {
		return
	}
	debitRegisterCount, err := getInt(line[19:27])
	if err != nil {
		return
	}
	totalRegisterCount, err := getInt(line[27:37])
	if err != nil {
		return
	}
	if math.Abs(totalAmount-p.doc.TotalAmount) > 0.01 {
		return fmt.Errorf("Calculated amount = %f diferent from parsed amount = %f", p.doc.TotalAmount, totalAmount)
	}
	p.doc.TotalAmount = totalAmount
	if debitRegisterCount != p.doc.DebitRegisterCount {
		return fmt.Errorf("Debit transactions on totals line = %d. Parsed debit transactions = %d", debitRegisterCount, p.doc.DebitRegisterCount)
	}
	p.countRegister()
	if totalRegisterCount != p.doc.TotalRegisterCount {
		return fmt.Errorf("Total doc register (99) missmatch. Calculated value = %d. Parsed = %d", totalRegisterCount, p.doc.TotalRegisterCount)
	}
	return
}

func getString(rs []rune) string {
	return strings.TrimSpace(string(rs))
}
func getDate(rs []rune) (date time.Time, err error) {
	date, err = time.Parse("20060102", getString(rs))
	return
}
func getMoney(rs []rune) (amount float64, err error) {
	amount, err = strconv.ParseFloat(getString(rs), 32)
	amount = amount / 100
	return
}
func getInt(rs []rune) (num int, err error) {
	num, err = strconv.Atoi(getString(rs))
	return
}
