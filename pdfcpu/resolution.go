package pdfcpu

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type NitInRut struct {
	NIT string
	DV  string
}

type Resolution struct {
	Prefix     string
	Number     int64
	LifeMonths int32
	StartDate  time.Time
	EndDate    time.Time
}

type PartialResolutionData struct {
	Resolution    Resolution
	InvoiceNumber int32
	InvoiceLimit  int32
}

func ExtractFileContent(filePath, tempDir, fileName string) (extractedLines string, patternsCount int, err error) {
	selectedPages := []string{"1", "2"}

	// Open the file as io.ReadSeeker
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("Error opening file:", err)
		return "", 0, err
	}

	// No configuration needed
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed

	// Extract content using the specified tempDir as outDir
	err = api.ExtractContent(file, tempDir, fileName, selectedPages, conf)
	file.Close() // Ensure the file is closed after processing

	if err != nil {
		log.Println("Error extracting content:", err)
		return "", 0, err
	}

	// Path to the extracted text files (adjust filename if necessary)
	var combinedContent strings.Builder
	for _, page := range selectedPages {
		txtPath := filepath.Join(tempDir, fmt.Sprintf("%s_Content_page_%s.txt", fileName, page))

		// Read the extracted content
		extractedContent, err := os.Open(txtPath)
		if err != nil {
			log.Println("Error reading extracted content:", err)
			return "", 0, err
		}

		// Convert extracted content to string
		decoder := transform.NewReader(extractedContent, charmap.Windows1252.NewDecoder())
		decoded, err := io.ReadAll(decoder)
		extractedContent.Close()

		if err != nil {
			log.Println("Error decoding content:", err)
			return "", 0, err
		}

		// Append decoded content to the combinedContent
		combinedContent.Write(decoded)
	}

	// Process combined content
	combinedText := combinedContent.String()
	re := regexp.MustCompile(`\(([^)]*)\)`)
	matches := re.FindAllStringSubmatch(combinedText, -1)
	var distilledContent strings.Builder

	for _, match := range matches {
		// Check for empty string or space
		if match[1] != " " && match[1] != "" {
			distilledContent.WriteString(fmt.Sprintf("%s\n", match[1]))
			patternsCount++
		}
	}

	// Return the distilled content and the count of matches
	return distilledContent.String(), patternsCount, nil
}

func ExtractIdentification(content string) (nitInRut *NitInRut, remainingLines []string) {
	lines := strings.Split(content, "\n")
	longNumberRegex := regexp.MustCompile(`^\d{10,}$`) // Pattern to identify long numbers
	letterLineRegex := regexp.MustCompile(`[a-zA-Z]`)  // Pattern to identify lines with letters

	// Find the first long number and slice from the line after it
	startIndex := -1
	for i, line := range lines {
		if longNumberRegex.MatchString(line) {
			startIndex = i + 1
			break
		}
	}

	// Find the first line with letters and slice up to it
	endIndex := len(lines)
	for i := startIndex; i < len(lines); i++ {
		if letterLineRegex.MatchString(lines[i]) {
			endIndex = i
			break
		}
	}

	// Extract relevant lines and determine identification and DV
	extractedLines := lines[startIndex:endIndex]
	dv := extractedLines[len(extractedLines)-1]
	identification := strings.Join(extractedLines[:len(extractedLines)-1], "")

	nitInRut = &NitInRut{
		NIT: identification,
		DV:  dv,
	}

	// if endIndex > 10 && endIndex < 20 && startIndex != -1 {
	// 	remainingLines = strings.Split(lines[0][:4], "")
	// } else {
	remainingLines = lines[endIndex+3:]
	// }

	return nitInRut, remainingLines
}

func ExtractResolutionFields(remainingLines []string) *PartialResolutionData {
	// Define the regex pattern for valid dates and times
	pattern := `(19|20)\d{2}-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01]) / (0[1-9]|1[0-2]):([0-5][0-9]):([0-5][0-9]) [AP]M`

	// Compile the regex
	re := regexp.MustCompile(pattern)

	var resolutionStartDate time.Time
	var resolutionNumber int64

	for i, line := range remainingLines {
		if re.MatchString(line) {
			// Parse StartDate from string to time.Time
			dateStr := line
			var err error
			if resolutionStartDate, err = time.Parse("2006-01-02 / 03:04:05 PM", dateStr); err != nil {
				fmt.Println("Error parsing date:", err)
				return nil
			}
			if i+1 < len(remainingLines) {
				numberStr := remainingLines[i+1]
				// Convert string to int64
				if resolutionNumber, err = strconv.ParseInt(numberStr, 10, 64); err != nil {
					fmt.Println("Error parsing number:", err)
					return nil
				}
			}
			break
		}
	}

	// Backward iteration to find the required fields
	var resolutionPrefix string
	var invoiceNumber int32
	var invoiceLimit int32
	var resolutionLifeMonths int32

	for i := len(remainingLines) - 1; i >= 0; i-- {
		if remainingLines[i] == "FACTURA ELECTRÓNICA DE VENTA" {
			if i+5 < len(remainingLines) {
				resolutionPrefix = remainingLines[i+2]
				var number int64
				var err error
				// Convert strings to int32
				if number, err = strconv.ParseInt(remainingLines[i+3], 10, 32); err == nil {
					invoiceNumber = int32(number)
				}
				if number, err = strconv.ParseInt(remainingLines[i+4], 10, 32); err == nil {
					invoiceLimit = int32(number)
				}
				if _, err = fmt.Sscanf(remainingLines[i+7], "%d", &resolutionLifeMonths); err != nil {
					fmt.Println("Error parsing life months:", err)
				}
			}
			break
		}
	}

	// Calculate EndDate by adding LifeMonths to StartDate
	endDate := resolutionStartDate.AddDate(0, int(resolutionLifeMonths), 0)

	resolution := &Resolution{
		Number:     resolutionNumber,
		Prefix:     resolutionPrefix,
		StartDate:  resolutionStartDate,
		EndDate:    endDate,
		LifeMonths: resolutionLifeMonths,
	}

	result := &PartialResolutionData{
		Resolution:    *resolution,
		InvoiceNumber: invoiceNumber,
		InvoiceLimit:  invoiceLimit,
	}

	return result
}
