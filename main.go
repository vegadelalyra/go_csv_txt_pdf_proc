package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/vegadelalyra/go_csv_txt_pdf_proc/pdfcpu"
)

func main() {
	// pdfcpu.ExtractRawPDF("data/example1.pdf", "RUT")
	// pdfcpu.ProcessExtractedPDF("RUT_Content_page_1.txt")
	// pdfcpu.ProcessExtractedPDF("output_Content_page_1.txt")

	t, n, _ := pdfcpu.ExtractFileContent("data/RESOLUCIÓN.pdf", "", "Resolución")
	i, r := pdfcpu.ExtractIdentification(t)
	z := pdfcpu.ExtractResolutionFields(r)

	zJSON, err := json.MarshalIndent(z, "", "  ")
	if err != nil {
		log.Fatalf("Error converting to JSON: %v", err)
	}

	fmt.Printf("number of patterns:\n%v\n", n)
	fmt.Printf("identification:\n%v\n", i)
	// fmt.Printf("remaining lines:\n %v\n", r)
	fmt.Printf("THE SCRAPED DATA: %s\n", zJSON)
}
