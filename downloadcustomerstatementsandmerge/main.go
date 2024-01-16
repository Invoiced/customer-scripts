package main

import (
	"fmt"
	"github.com/Invoiced/invoiced-go/v2/api"
	"github.com/signintech/gopdf"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Config represents the structure of the configuration file
type Config struct {
	StartDate     string `yaml:"start_date"`
	EndDate       string `yaml:"end_date"`
	Sandbox       bool   `yaml:"sandbox"`
	APIKey        string `yaml:"api_key"`
	StatementType string `yaml:"statement_type"`
}

func main() {
	// Read the configuration file
	configFile, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// Unmarshal the YAML into our Config struct
	var config Config
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}

	// Parse the start date
	startDate, err := parseDateTime(config.StartDate)
	if err != nil {
		log.Fatalf("Error parsing start date: %v", err)
	}

	// Parse the end date
	endDate, err := parseDateTime(config.EndDate)
	if err != nil {
		log.Fatalf("Error parsing end date: %v", err)
	}

	// Print out the configuration values
	fmt.Println("Start Date:", startDate)
	fmt.Println("End Date:", endDate)
	fmt.Println("Sandbox:", config.Sandbox)
	fmt.Println("API Key:", config.APIKey)

	client := api.New(config.APIKey, config.Sandbox)

	customers, err := client.Customer.ListAll(nil, nil)

	if err != nil {
		fmt.Println("Error getting customers", err)
		return
	}

	var pdfFiles []string

	for _, customer := range customers {
		balance, err := client.Customer.GetBalance(customer.Id)
		if err != nil {
			fmt.Println("Error getting balance for customer", customer.Id, err)
			continue
		}

		if balance.DueNow == 0 {
			continue
		}

		pdfStatement := customer.StatementPdfUrl

		pdfPath, err := fetchPDF(pdfStatement, strconv.FormatInt(customer.Id, 10), startDate, endDate)

		if err != nil {
			panic(err)
		}

		fmt.Println("Downloaded PDF for customer", customer.Number)

		pdfFiles = append(pdfFiles, pdfPath)
	}

	err = stitchPDFs(pdfFiles, "statement_to_print.pdf")

	if err != nil {
		log.Fatalf("Error stitching PDFs: %v", err)
	} else {
		fmt.Println("Stitched PDFs successfully")
		deleteFiles(pdfFiles)
	}

}

func fetchPDF(url, customerId, statementType string, startDate, endDate time.Time) (string, error) {
	if statementType == "balance_forward" {
		url += fmt.Sprintf("?statement_type=%s&&start=%d&end=%d", statementType, startDate.Unix(), endDate.Unix())
	} else if statementType == "open_item" {
		url += fmt.Sprintf("?statement_type=%s&end=%d", statementType, endDate.Unix())
	}

	fmt.Println("Fetching PDF from", url, "for customer", customerId)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	filePath := fmt.Sprintf("statement_%s_%s_%s.pdf", customerId, startDate.Unix(), endDate.Unix())
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func stitchPDFs(filePaths []string, outputFileName string) error {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	for _, filePath := range filePaths {
		pdf.AddPage()
		tpl := pdf.ImportPage(filePath, 1, "/MediaBox")
		pdf.UseImportedTemplate(tpl, 0, 0, gopdf.PageSizeA4.W, gopdf.PageSizeA4.H)
	}

	err := pdf.WritePdf(outputFileName)
	if err != nil {
		return err
	}

	return nil
}

func deleteFiles(filePaths []string) {
	for _, filePath := range filePaths {
		err := os.Remove(filePath)
		if err != nil {
			log.Printf("Failed to delete file %s: %v\n", filePath, err)
		} else {
			log.Printf("Deleted file %s\n", filePath)
		}
	}
}

func parseDateTime(dateTimeStr string) (time.Time, error) {
	layout := "2006-01-02 03:04 PM MST"
	return time.Parse(layout, dateTimeStr)
}
