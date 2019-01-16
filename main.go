package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/apsl/sepakit/convert"
)

func main() {
	inpath := "-"
	outpath := "-"
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Converts AEB 19.14 TXT file to SEPA XML file\nUsage: %s [INFILE] [OUTFILE]\nDefaults to stdin and stdout (-)\n", os.Args[0])
	}
	flag.Parse()

	if len(os.Args) > 1 {
		inpath = os.Args[1]
	}
	if len(os.Args) > 2 {
		outpath = os.Args[2]
	}

	var err error
	fin := os.Stdin
	if inpath != "-" {
		fin, err = os.Open(inpath)
		if err != nil {
			log.Fatalf("%s: file not found\n", inpath)
		}
		defer fin.Close()
	}

	fout := bufio.NewWriter(os.Stdout)
	if outpath != "-" {
		f, err := os.Create(outpath)
		if err != nil {
			log.Fatalf("Cannot open file %s for writing: %s\n", outpath, err)
		}
		defer f.Close()
		fout = bufio.NewWriter(f)
	}

	err = convert.Latin1DebitTxtToXML(fin, fout)
	if err != nil {
		log.Fatal("error writting xml: ", err)
	}
	fout.Flush()
}
