package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// Export Customer With Magic Login Links
// This script prints CSV-formatted data to stdout in the following format:
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
)

type invoicedCustomer struct {
	Name   string `json:"name"`
	Number string `json:"number"`
	ID     int    `json:"id"`
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

	// print first row for CSV format
	fmt.Println("Customer Name,Customer Number,Magic Link")

	customers, err := client.getAllCustomers()
	if err != nil {
		panic(err.Error())
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

		row := c.Name + "," + c.Number + "," + "https://" + InvoicedCompanySubdomain +
			".invoiced.com/login/" + tokenString

		fmt.Println(row)
	}
}
