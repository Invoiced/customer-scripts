package main

import (
	"flag"
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
	StartDate       string `yaml:"start_date"`
	EndDate         string `yaml:"end_date"`
	Sandbox         bool   `yaml:"sandbox"`
	APIKey          string `yaml:"api_key"`
	StatementType   string `yaml:"statement_type"`
	Currency        string `yaml:"currency"`
	SkipZeroBalance bool   `yaml:"skip_zero_balance"`
}

func main() {

	timeBegin := time.Now()

	configFileName := flag.String("config", "config.yaml", "specify the config file to use")
	flag.Parse()
	// Read the configuration file
	configFile, err := os.ReadFile(*configFileName)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// Unmarshal the YAML into our Config struct
	var config Config
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}

	fmt.Println("Start Date: ", config.StartDate)
	fmt.Println("End Date: ", config.EndDate)
	fmt.Println("Statement Currency: ", config.Currency)
	fmt.Println("Statement Type: ", config.StatementType)

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

	client := api.New(config.APIKey, config.Sandbox)

	customers, err := client.Customer.ListAll(nil, nil)

	if err != nil {
		fmt.Println("Error getting customers", err)
		return
	}

	var pdfFiles []string

	statementCounter := 0

	for _, customer := range customers {

		if config.SkipZeroBalance {
			balance, err := client.Customer.GetBalance(customer.Id)
			if err != nil {
				fmt.Println("Error getting balance for customer", customer.Id, err)
				continue
			}

			if balance.TotalOutstanding == 0 {
				continue
			}
		}

		pdfStatement := customer.StatementPdfUrl

		pdfPath, err := fetchPDF(pdfStatement, config.StatementType, strconv.FormatInt(customer.Id, 10), config.Currency, startDate, endDate)

		if err != nil {
			panic(err)
		}

		pdfFiles = append(pdfFiles, pdfPath)
		statementCounter++
	}

	statementName := fmt.Sprintf("customer_statements_%d_%d.pdf", startDate.Unix(), endDate.Unix())
	err = stitchPDFs(pdfFiles, statementName)

	if err != nil {
		log.Fatalf("Error stitching PDFs: %v", err)
	} else {
		fmt.Println("Stitched PDFs successfully")
		deleteFiles(pdfFiles)
	}

	timeEnd := time.Now()
	timeTake := timeEnd.Sub(timeBegin)

	if timeTake.Minutes() < 1 {
		fmt.Println("Merged", statementCounter, "customer statements in", timeEnd.Sub(timeBegin).Seconds(), "seconds")
	} else {
		fmt.Println("Merged", statementCounter, "customer statements in", timeEnd.Sub(timeBegin).Minutes(), "minutes")
	}

}

func fetchPDF(url, statementType, customerId, currency string, startDate, endDate time.Time) (string, error) {

	if statementType == "balance_forward" {
		url += fmt.Sprintf("?statement_type=%s&&start=%d&end=%d&currency=%s", statementType, startDate.Unix(), endDate.Unix(), currency)
	} else if statementType == "open_item" {
		url += fmt.Sprintf("?statement_type=%s&end=%d&currency=%s&items=open", statementType, endDate.Unix(), currency)
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	filePath := fmt.Sprintf("statement_%s_%d_%d.pdf", customerId, startDate.Unix(), endDate.Unix())
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
		}
	}
}

func parseDateTime(dateTimeStr string) (time.Time, error) {
	layout := "2006-01-02 03:04:05 PM MST"
	return time.Parse(layout, dateTimeStr)
}
