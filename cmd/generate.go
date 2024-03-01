package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
)

var paymentType string

const sheetName = "Invoice Template"

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates a new invoice based on the specified payment type",
	Long: `Generate is a command to create an invoice with specific details based on the payment type.
Example usage:

./fastinvoice generate --payment-type ach`,
	Run: func(cmd *cobra.Command, args []string) {
		generateInvoice(paymentType)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().StringVarP(&paymentType, "payment-type", "p", "", "Type of payment (ach, domestic wire, international wire, check)")
	generateCmd.MarkFlagRequired("payment-type")
}

func generateInvoice(paymentType string) {
	f, err := excelize.OpenFile("template.xlsx")
	if err != nil {
		log.Fatalf("Failed to open template: %v", err)
	}

	// Vendor Information
	f.SetCellValue(sheetName, "B2", "Acme Corp")
	f.SetCellValue(sheetName, "B3", "123 Street Address, City, State, Zip")
	f.SetCellValue(sheetName, "B4", "www.acmecorp.com, info@acmecorp.com")
	f.SetCellValue(sheetName, "B5", "+1234567890")

	// Billed To Information
	f.SetCellValue(sheetName, "B10", "John Doe")
	f.SetCellValue(sheetName, "B11", "Client Company Name")
	f.SetCellValue(sheetName, "B12", "Client Address")
	f.SetCellValue(sheetName, "B13", "Client Phone, Client Email")

	// Ship To Information
	f.SetCellValue(sheetName, "D10", "Jane Doe / Dept")
	f.SetCellValue(sheetName, "D11", "Client Company Name")
	f.SetCellValue(sheetName, "D12", "Client Address")
	f.SetCellValue(sheetName, "D13", "Client Phone")

	// Invoice Metadata
	f.SetCellValue(sheetName, "F9", "#INV00001")
	f.SetCellValue(sheetName, "F10", "11/11/11")
	f.SetCellValue(sheetName, "F11", "12/12/12")

	// Example Line Item
	f.MergeCell(sheetName, "B16", "C16")
	f.SetCellValue(sheetName, "B16", "Service Name")
	f.SetCellValue(sheetName, "D16", 2)      // Quantity
	f.SetCellValue(sheetName, "E16", 500.00) // Unit Price
	// Assuming F16 has a formula to calculate the total

	// Payment Instructions and Terms
	f.MergeCell(sheetName, "B36", "F36")
	f.SetCellValue(sheetName, "B36", "Payment instructions here, e.g., bank, PayPal...")
	f.MergeCell(sheetName, "B37", "F37")
	f.SetCellValue(sheetName, "B37", "Terms here, e.g., warranty, returns policy...")

	// Save the filled-out template to a new file
	if err := f.SaveAs("filled_invoice.xlsx"); err != nil {
		log.Fatalf("Failed to save filled invoice: %v", err)
	}
	fmt.Println("Invoice generated successfully")
}
