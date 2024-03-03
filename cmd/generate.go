package cmd

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jaswdr/faker/v2"
	"github.com/lucasheriques/fastinvoice/utils"
	"github.com/spf13/cobra"
)

type InvoiceData struct {
	CompanyLogo    string
	InvoiceNumber  string
	InvoiceDate    string
	DueDate        string
	VendorInfo     string
	CustomerInfo   string
	PaymentMethod  string
	PaymentDetails string
	Items          []InvoiceItem
	Total          string
}

type InvoiceItem struct {
	Description string
	Price       string
}

var paymentMethod string

const tmplFile = "invoice.tmpl"
const convertAPIURL = "http://127.0.0.1:52171/convert/html"

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates a new invoice based on the specified payment type",
	Long: `Generate is a command to create an invoice with specific details based on the payment type.
Example usage:

./fastinvoice generate --payment-method ach`,
	Run: func(cmd *cobra.Command, args []string) {
		generateInvoice(paymentMethod)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringVarP(&paymentMethod, "payment-type", "p", "", "Type of payment (ach, domestic wire, international wire, check)")
	generateCmd.MarkFlagRequired("payment-method")
}

func generateData(paymentMethod string) InvoiceData {
	fake := faker.New()
	companyName := fake.Company().Name()
	now := time.Now()
	companyAddress := fake.Company().Faker.Address()

	companyEmail := "bills@" + utils.TransformIntoValidEmailName(companyName) + "." + fake.Internet().Domain()

	data := InvoiceData{
		CompanyLogo: "https://example.com/logo.png",
		// convert from int to string
		InvoiceNumber: strconv.Itoa(fake.RandomNumber(5)),
		// Invoice date should be today's date
		InvoiceDate: now.Format("January 2, 2006"),
		// Due date should be 30 days from today
		DueDate: now.AddDate(0, 0, 30).Format("January 2, 2006"),
		VendorInfo: fmt.Sprintf(`%s
		%s
		%s %s
		%s`, companyName, companyAddress.StreetAddress(), companyAddress.StateAbbr(), companyAddress.SecondaryAddress(), companyEmail),
		CustomerInfo: `Acme Corp.
		John Doe
		john@example.com`,
		PaymentMethod:  "ACH",
		PaymentDetails: "Routing number: 026001591. Account number: 7534028150001. Beneficiary name: TechWave Solutions",
		Items: []InvoiceItem{
			{Description: "Website design", Price: "$300.00"},
			{Description: "Hosting (3 months)", Price: "$75.00"},
			{Description: "Domain name (1 year)", Price: "$10.00"},
		},
		Total: "$385.00",
	}

	return data
}

func convertAndDownloadPdf(filePath string) {
	htmlContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Error reading HTML file: %v", err)
	}
	defer os.Remove(filePath)

	// Create a new request with the HTML content
	response, err := http.Post(convertAPIURL, "text/html", bytes.NewReader(htmlContent))
	if err != nil {
		log.Fatalf("Error sending request to the API: %v", err)
	}
	defer response.Body.Close()

	// Check the response
	if response.StatusCode != http.StatusOK {
		log.Fatalf("Failed to convert HTML to PDF, API responded with status code: %d", response.StatusCode)
	}

	// Create the PDF file
	pdfFile, err := os.Create("output.pdf")
	if err != nil {
		log.Fatalf("Error creating PDF file: %v", err)
	}
	defer pdfFile.Close()

	// Copy the response body (PDF content) to the PDF file
	_, err = io.Copy(pdfFile, response.Body)
	if err != nil {
		log.Fatalf("Error saving PDF file: %v", err)
	}

	log.Println("PDF successfully created and saved as output.pdf")
}

func generateHtmlFile(invoiceData InvoiceData) string {
	templ, err := template.New(tmplFile).Funcs(template.FuncMap{
		"nl2br": func(text string) template.HTML {
			return template.HTML(strings.Replace(html.EscapeString(text), "\n", " <br/> ", -1))
		},
	}).ParseFiles(tmplFile)
	if err != nil {
		log.Println("Error parsing template")
		log.Fatal(err)
	}

	file, err := os.Create("index.html")
	if err != nil {
		log.Println("Error creating file")
		log.Fatal(err)
	}
	defer file.Close()

	err = templ.Execute(file, invoiceData)
	if err != nil {
		log.Println("Error executing template")
		log.Fatal(err)
	}

	return file.Name()
}

func generateInvoice(paymentMethod string) {
	invoiceData := generateData(paymentMethod)

	htmlFile := generateHtmlFile(invoiceData)

	convertAndDownloadPdf(htmlFile)
}
