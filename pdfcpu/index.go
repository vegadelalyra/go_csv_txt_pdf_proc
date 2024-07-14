package pdfcpu

import (
	"fmt"
	"os"
	"regexp"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func ExtractRawPDF(pdfPath string) {
	// Open the PDF file
	pdfFile, err := os.Open(pdfPath)
	if err != nil {
		fmt.Println("Error opening PDF file:", err)
		return
	}
	defer pdfFile.Close() // Close the file after usage

	// No specific output directory needed (set to empty string)
	outDir := ""

	// No specific filename needed (set to empty string)
	fileName := "RUT"

	// Empty slice for selectedPages (extract current page)
	selectedPages := []string{"1"}

	// No configuration needed
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed

	// Extract content using the correct import path
	err = api.ExtractContent(pdfFile, outDir, fileName, selectedPages, conf)
	if err != nil {
		fmt.Println("Error extracting content:", err)
		return
	}
}

func ProcessExtractedPDF(txtPath string) {
	// Read the extracted content (assuming a single file)
	extractedContent, err := os.ReadFile(txtPath) // Adjust filename if different
	if err != nil {
		fmt.Println("Error reading extracted content:", err)
		return
	}

	// Convert extracted content to string
	contentString := string(extractedContent)
	re := regexp.MustCompile(`\(([^)]*)\)`)

	// Find all matches
	matches := re.FindAllStringSubmatch(contentString, -1)

	i := 1
	for _, match := range matches {
		// Check for empty string or space
		if match[1] != " " && match[1] != "" {
			// Format the line with index
			fmt.Printf("%d: %s\n", i, match[1])
			i++ // Increment index after printing
		}
	}
}
