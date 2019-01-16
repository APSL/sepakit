# sepakit 

Converts AEB 19.14 Direct Debit TXT file to SEPA XML (Core scheme)

[![Go Report Card](https://goreportcard.com/badge/github.com/apsl/sepakit)](https://goreportcard.com/report/github.com/apsl/sepakit)

sepakit implements a small golang library and command line tool that can parse a **AEB 19-14** direct debit text file, and generate a SEPA  ISO-20022 XML **PAIN.008.001.02** (Core scheme) direct debit file.

*Motivation*: European banks no longer admit Direct Debit files in plain text AEB 19.14 format, and some legacy programs do not provide means to transform its output to the new European SEPA norm.

Command sepakit takes filenames or stdin/stdout as arguments and converts from AEB 1419 to SEPA XML.
File [input-aeb1914.txt](input-aeb1914.txt) is provided as an example of input AEB 1914 text file.


Can be used as a library:

* package *aeb1914* implements the AEB-1914 parser. Accepts io.Reader as input parameter.
* package *sepadebit* implements the SEPA XML writer. Outputs to an io.Writer.
* package *convert* uses the former ones and do the whole parsing and XML generation. 

**note**: Eeach country and bank has its particularities on SEPA usage. This tool was coded for converting old aeb format to the SEPA XML used by CAIXABANK spanish banking entity, who requires iso-8859-1 format, and a hard restriction on its allowed characters. Not tested for other banks/countries, but should work with some modifcation.
Some Creditor and payment optional parameters have been suited following CAIXABANK SEPA manuals. BIC is hardcoded (a BIC generator should be implemented). See [convert.go](convert/convert.go))

See also http://github.com/bercab/txp for a simple convert desktop utility (linux and windows) using this package.


## Install

```
go get github.com/apsl/sepakit
```

## Example command line usage

```
sepakit input-aeb1914.txt out.xml
```



## Exampe package aeb1914 usage 

```go
    import "github.com/apsl/sepakit/aeb1914"

    parser := NewParser()
    r, err := os.Open("input-aeb1914.txt")
    doc, err := parser.Parse(r)
    fmt.Printf("%s\n", doc.InitiatingParty.Name)
    fmt.Printf("%f\n", doc.TotalAmount)

```

## Example sepadebit usage


```go
    import "github.com/apsl/sepakit/sepadebit"

    docxml := sepadebit.NewDocument()
    docxml.SetInitiatingParty(initiatingPartyName, creditorID)
    //[...]
    w := os.Stdout
    err = docxml.WriteLatin1(w)
```

See  DebitTxtToXML on convert package for a full sepadebit.Document generation.

## Example full convert usage

```go
    import "github.com/apsl/sepakit/convert"

    r := os.Stdin
    w := os.Stdout
    err = convert.Latin1DebitTxtToXML(r, w)
```

## Other SEPA ISO-20022 XML golang packages

Apart from Direct Debit transfers, there are other SEPA estandard documents not covered here.

See also [gosepa](https://github.com/Softinnov/gosepa) for a golang pain.002.001.03 schema (Customer Credit Transfer Initiation V03) file generator.

## Author

Bernardo Cabezas - bcabezas at [apsl.net](https://www.apsl.net) - [bercab](https://github.com/bercab)

## License

This project is licensed under the MIT License - see the [LICENSE.txt](LICENSE.txt) file for details