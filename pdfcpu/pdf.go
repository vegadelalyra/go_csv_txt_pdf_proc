package pdfcpu

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type PartialRutData struct {
	FirstName     string
	SecondName    string
	FirstSurname  string
	SecondSurname string
	CompanyName   string
	PartyType     string
	IDType        string
	CountryCode   string
	Dept          string
	City          string
	Address       string
	Email         string
	Phone1        string
	Phone2        string
	TaxLevel      string
}

func ExtractRawPDF(pdfPath, fileName string) {
	// Open the PDF file
	pdfFile, err := os.Open(pdfPath)
	if err != nil {
		fmt.Println("Error opening PDF file:", err)
		return
	}
	defer pdfFile.Close() // Close the file after usage

	// No specific output directory needed (set to empty string)
	outDir := ""

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
	extractedContent, err := os.Open(txtPath) // Adjust filename if different
	if err != nil {
		fmt.Println("Error reading extracted content:", err)
		return
	}
	defer extractedContent.Close()

	// Convert extracted content to string
	decoder := transform.NewReader(extractedContent, charmap.Windows1252.NewDecoder())
	decoded, err := io.ReadAll(decoder)
	if err != nil {
		fmt.Println("Error decoding content:", err)
		return
	}

	contentString := string(decoded)

	// Print the content
	re := regexp.MustCompile(`\(([^)]*)\)`)

	// Find all matches
	matches := re.FindAllStringSubmatch(contentString, -1)
	var extractedText strings.Builder
	for _, match := range matches {
		// Check for empty string or space
		if match[1] != " " && match[1] != "" {
			extractedText.WriteString(fmt.Sprintf("%s\n", match[1]))
		}
	}

	id, dv, remainingLines := extractIdentification(extractedText.String())

	fmt.Printf("id: %v\n", id)
	fmt.Printf("dv: %v\n", dv)
	fmt.Printf("contentString: %v\n", contentString)

	extractRemainingFields(remainingLines)
}

func extractIdentification(content string) (identification, dv string, remainingLines []string) {
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
	dv = extractedLines[len(extractedLines)-1]
	identification = strings.Join(extractedLines[:len(extractedLines)-1], "")
	remainingLines = lines[endIndex+3:]

	return identification, dv, remainingLines
}

func extractRemainingFields(remainingLines []string) *PartialRutData {
	partyType := extractPartyType(remainingLines[0])
	firstSurname, secondSurname, firstName, secondName, remainingLines := extractName(remainingLines)
	companyName, remainingLines := extractCompanyName(remainingLines)
	department, city, address, email, remainingLines := extractAddressAndEmail(remainingLines)
	taxLevel, remainingLines := extractTaxLevel(remainingLines)
	phone1, phone2 := extractPhones(remainingLines)

	extractedData := PartialRutData{
		FirstSurname:  firstSurname,
		SecondSurname: secondSurname,
		FirstName:     firstName,
		SecondName:    secondName,
		CompanyName:   companyName,
		PartyType:     partyType,
		Dept:          department,
		City:          city,
		Address:       address,
		Email:         email,
		Phone1:        phone1,
		Phone2:        phone2,
		TaxLevel:      taxLevel,
	}
	return &extractedData
}

func skipZipCode(remainingLines []string) []string {
	remainingLines = remainingLines[9:]

	if remainingLines[0] != "3" {
		remainingLines = remainingLines[6:]
	}

	return remainingLines
}

// Extract Party Type from client's uploaded RUT
func extractPartyType(line string) string {

	if strings.Contains(strings.ToLower(line), "natural") {
		return "PERSONA_NATURAL"
	} else if strings.Contains(strings.ToLower(line), "jurídica") {
		return "PERSONA_JURIDICA"
	}
	return line // Return the original line if no keywords matched
}

// Extract Name Complete from client's uploaded RUT
func extractName(lines []string) (firstSurname, secondSurname, firstName, secondName string, remainingLines []string) {
	letterPattern := regexp.MustCompile(`[A-Za-z]`)
	counter := 0

	for i, line := range lines {
		if letterPattern.MatchString(line) {
			counter++
			if counter == 6 {
				remainingLines = lines[i:]
				break
			}
		}
	}

	firstSurname = remainingLines[0]
	secondSurname = remainingLines[1]
	firstName = remainingLines[2]
	secondName = remainingLines[3]

	return firstSurname, secondSurname, firstName, secondName, remainingLines
}

// Extract Company Name if applies from client's uploaded RUT
func extractCompanyName(lines []string) (companyName string, remainingLines []string) {
	jumpToLocationLines := 8
	if lines[4] != "COLOMBIA" {
		companyName = lines[4]
		jumpToLocationLines++
	}
	remainingLines = lines[jumpToLocationLines:]

	return companyName, remainingLines
}

// Extract address and email from client's uploaded RUT
func extractAddressAndEmail(lines []string) (department, city, address, email string, remainingLines []string) {
	department = lines[0]
	city = lines[3]
	address = lines[7]
	email = lines[8]

	remainingLines = skipZipCode(lines)

	return department, city, address, email, remainingLines
}

// Extract Tax Type from client's uploaded RUT
func extractTaxLevel(lines []string) (taxLevel string, remainingLines []string) {
	aDigitAHyphenALetter := regexp.MustCompile(`^\d+\s*-\s*.*$`)
	for i, line := range lines {
		if aDigitAHyphenALetter.MatchString(line) {
			if strings.Contains(line, "48") {
				taxLevel = "COMUN"
			} else {
				taxLevel = "SIMPLIFICADO"
			}
			remainingLines = lines[:i]
			break
		}
	}

	return taxLevel, remainingLines
}

// Extract Phone1 and Phone2 from client's uploaded RUT
func extractPhones(remainingLines []string) (phone1, phone2 string) {
	phones := strings.Join(remainingLines, "")
	fourRanDigitsADate := regexp.MustCompile(`\d{4}(19[0-9]{2}|20[0-9]{2})(0[1-9]|1[0-2])(0[1-9]|[12][0-9]|3[01])`)

	matchIndex := fourRanDigitsADate.FindStringIndex(phones)
	phones = phones[:matchIndex[0]]

	if len(phones) == 7 || len(phones) == 10 {
		phone1 = phones
	} else if len(phones) == 20 {
		phone1 = phones[:10]
		phone2 = phones[10:]
	} else if len(phones) == 14 {
		phone1 = phones[:7]
		phone2 = phones[7:]
	} else if len(phones) == 17 {
		firstThreeDigits, err := strconv.Atoi(phones[:3])
		if err == nil && firstThreeDigits >= 300 && firstThreeDigits <= 324 && phones[10] != '3' {
			phone1 = phones[:10]
			phone2 = phones[10:]
		} else {
			phone1 = phones[:7]
			phone2 = phones[7:]
		}
	}

	return phone1, phone2
}

/*
	initial deadline
	identification
	verification digit
	pre party type deadline
	party type PERSONA_NATURAL || PERSONA_JURIDICA
	identification type

	prename deadline
	first surname
	second surname
	first name
	second name

	company name if applies

	country
	pre department deadline
	department
	pre city deadline
	city

	pre address deadline
	address
	email
	unused zip code COLOMBIANS ONES HAVE 6 DIGITS

	phone number one
	phone number two if applies

	pre tax type dead line
	taxtype NUMBER + HYPHEN SYMBOL + LETTERS if NUMBER=48 COMUN else SIMPLIFICADO
	dead end.


*/

/*
	upload pdf
	extract text
	distill data
	scrape id + dv fields
	validate id === session.id || id + dv === session.id+dv
	scrape remaining fields
	save in one batch all fields on db
*/
