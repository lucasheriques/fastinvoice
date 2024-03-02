package cmd

import (
	"bytes"
	"html"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jaswdr/faker/v2"
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
	now := time.Now()

	data := InvoiceData{
		CompanyLogo: "https://example.com/logo.png",
		// convert from int to string
		InvoiceNumber: strconv.Itoa(fake.RandomNumber(5)),
		// Invoice date should be today's date
		InvoiceDate: now.Format("January 2, 2006"),
		// Due date should be 30 days from today
		DueDate: now.AddDate(0, 0, 30).Format("January 2, 2006"),
		VendorInfo: `TechWave Solutions
		8 10th St San Francisco
		CA 94103
		invoices@faketechwave.com`,
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

func convertAndDownloadPdf(htmlFilePath string) {
	// URL of the Gotenberg API endpoint for converting HTML to PDF
	apiURL := "http://localhost:3000/forms/chromium/convert/html"

	// Prepare the file to be uploaded
	file, err := os.Open(htmlFilePath)
	if err != nil {
		log.Fatalf("Error opening HTML file: %v", err)
	}
	defer file.Close()

	// Prepare a form that you will submit to the endpoint
	var requestBody bytes.Buffer
	multiPartWriter := multipart.NewWriter(&requestBody)

	// Add the HTML file to the form
	fileWriter, err := multiPartWriter.CreateFormFile("files", filepath.Base(htmlFilePath))
	if err != nil {
		log.Fatalf("Error adding HTML file to form: %v", err)
	}
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		log.Fatalf("Error copying HTML file to form: %v", err)
	}

	// Important: Close the multipart writer to set the terminating boundary
	err = multiPartWriter.Close()
	if err != nil {
		log.Fatalf("Error closing multipart writer: %v", err)
	}

	// Create a new request to the API endpoint
	request, err := http.NewRequest("POST", apiURL, &requestBody)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// Set the content type to multipart/form-data followed by the boundary parameter
	request.Header.Set("Content-Type", multiPartWriter.FormDataContentType())

	// Perform the request
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Fatalf("Error sending request to Gotenberg API: %v", err)
	}
	defer response.Body.Close()

	// Check the response
	if response.StatusCode != http.StatusOK {
		log.Fatalf("Failed to convert HTML to PDF, Gotenberg API responded with status code: %d", response.StatusCode)
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

func generateInvoice(paymentMethod string) {
	invoiceData := generateData(paymentMethod)

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
	// Also delete the file once the function is done
	defer os.Remove(file.Name())

	err = templ.Execute(file, invoiceData)
	if err != nil {
		log.Println("Error executing template")
		log.Fatal(err)
	}

	convertAndDownloadPdf(file.Name())
}
