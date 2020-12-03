package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// Export Customer With Magic Login Links
// This script prints CSV-formatted data to excel in the following format:
// Customer Name | Customer Number | Magic Link
// string        | string          | URL (string)

const (
	baseURL = "https://api.invoiced.com"
	testURL = "https://api.sandbox.invoiced.com"

	InvoicedMagicLinkKey = "" // FIXME: Add Invoiced magic link key
	InvoicedApiKey       = "" // FIXME: Add Invoiced regular API Key

	InvoicedCompanySubdomain = ""   // FIXME: Add company subdomain like "acmeinc"; append ".sandbox" if test account
	IsInvoicedSandbox        = true // FIXME: Change to false if account is not sandbox

	TokenTTL = time.Hour * 24 * 90 // FIXME: Update token TTL as desired; default is 90 days

	ResultFileName = "output" // FIXME: Edit file name (not including .xlsx suffix) for exported file as desired
)

type invoicedCustomer struct {
	Name   string `json:"name"`
	Number string `json:"number"`
	ID     int    `json:"id"`
	Email  string `json:"email"`
}

type InvoicedClient struct {
	APIKey   string
	Http     *http.Client
	TestMode bool
}

func NewInvoicedClient(key string, test bool) *InvoicedClient {
	return &InvoicedClient{
		APIKey:   key,
		Http:     &http.Client{Timeout: 10 * time.Second},
		TestMode: test,
	}
}

func (ic *InvoicedClient) getAllCustomers() ([]invoicedCustomer, error) {
	base := baseURL
	if ic.TestMode {
		base = testURL
	}

	var output []invoicedCustomer
	page := 1

	for {
		var responseHolder []invoicedCustomer

		req, err := http.NewRequest(http.MethodGet, base+"/customers?page="+strconv.Itoa(page), nil)
		if err != nil {
			return nil, err
		}

		req.SetBasicAuth(ic.APIKey, "")

		resp, err := ic.Http.Do(req)
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()
		respBody, err := ioutil.ReadAll(resp.Body)
		statusCode := resp.StatusCode
		if err != nil {
			return nil, err
		}

		if statusCode >= http.StatusBadRequest {
			return nil, errors.New("Received " + strconv.Itoa(statusCode) + " response against request GET " + req.URL.String())
		}

		err = json.Unmarshal(respBody, &responseHolder)
		if err != nil {
			return nil, err
		}

		if len(responseHolder) == 0 {
			break
		}

		output = append(output, responseHolder...)
		page += 1
	}

	return output, nil
}

func main() {
	// set up invoiced client
	client := NewInvoicedClient(InvoicedApiKey, IsInvoicedSandbox)

	// create excel file and specify header rows
	f := excelize.NewFile()
	_ = f.SetCellValue("Sheet1", "A1", "Customer Name")
	_ = f.SetCellValue("Sheet1", "B1", "Customer Number")
	_ = f.SetCellValue("Sheet1", "C1", "Customer Email")
	_ = f.SetCellValue("Sheet1", "D1", "Magic Link")

	// set starting row number
	rowNum := 2

	customers, err := client.getAllCustomers()
	if err != nil {
		panic(err)
	}

	for _, c := range customers {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": strconv.Itoa(c.ID),
			"iss": "Invoiced Customer Script",
			"exp": time.Now().Add(TokenTTL).Unix(),
		})

		tokenString, err := token.SignedString([]byte(InvoicedMagicLinkKey))
		if err != nil {
			fmt.Println(err.Error())
		}

		_ = f.SetCellValue("Sheet1", "A"+strconv.Itoa(rowNum), c.Name)
		_ = f.SetCellValue("Sheet1", "B"+strconv.Itoa(rowNum), c.Number)
		_ = f.SetCellValue("Sheet1", "C"+strconv.Itoa(rowNum), c.Email)
		_ = f.SetCellValue("Sheet1", "D"+strconv.Itoa(rowNum),
			"https://"+InvoicedCompanySubdomain+".invoiced.com/login/"+tokenString)

		rowNum += 1
	}

	if err = f.SaveAs(ResultFileName+".xlsx"); err != nil {
		panic(err)
	}
}
