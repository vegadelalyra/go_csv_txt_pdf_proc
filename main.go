package main

import (
	"github.com/vegadelalyra/go_csv_txt_pdf_proc/pdfcpu"
)

func main() {
	// pdfcpu.ExtractRawPDF("data/example1.pdf")
	pdfcpu.ProcessExtractedPDF("output_Content_page_1.txt")
}
