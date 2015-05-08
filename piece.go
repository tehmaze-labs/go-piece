package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/tehmaze-labs/go-piece/parser"
	"github.com/tehmaze-labs/go-sauce"
)

func main() {
	flag.Parse()
	format := flag.String("format", "html", "Output format")

	for _, filename := range flag.Args() {
		f, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		w, h := 80, 25

		s, err := sauce.Parse(filename)
		if err != nil {
			log.Printf("%s: failed to parse SAUCE: %v\n", filename, err)
		}
		if s == nil {
			log.Printf("%s: no SAUCE record\n", filename)
		} else {
			if s.DataType == sauce.DATA_TYPE_CHARACTER {
				switch s.FileType {
				case 0, 1:
					w = int(s.TInfo[0])
				}
			}
		}

		log.Printf("creating %d x %d buffer\n", w, h)
		p := parser.NewANSI(w, h)
		p.Parse(f)

		switch *format {
		case "html":
			fmt.Println(p.Html())
		case "text":
			fmt.Println(p.String())
		default:
			log.Fatalf("Unknown format %q\n", *format)
		}
	}
}
