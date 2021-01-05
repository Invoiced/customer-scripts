package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// -- Delete Payments in Date Range --
// This script voids all payments within a certain date range.

const (
	baseURL = "https://api.invoiced.com"
	testURL = "https://api.sandbox.invoiced.com"

	dateRangeBegin = "2020-01-01"
	dateRangeEnd   = "2022-12-31"

	InvoicedApiKey    = ""   // FIXME: Add Invoiced API Key
	IsInvoicedSandbox = true // FIXME: Change to false if account is not sandbox
)

type payment struct {
	ID     int
	Date   int64
	Voided bool
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

func (ic *InvoicedClient) getAllPaymentsInRange(begin int64, end int64) ([]payment, error) {
	base := baseURL
	if ic.TestMode {
		base = testURL
	}

	var output []payment
	page := 1

	for {
		var responseHolder []payment

		req, err := http.NewRequest(http.MethodGet, base+"/payments?start_date="+strconv.FormatInt(begin, 10)+
			"&end_date="+strconv.FormatInt(end, 10)+"&page="+strconv.Itoa(page), nil)
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

func (ic *InvoicedClient) deletePayment(id int) error {
	base := baseURL
	if ic.TestMode {
		base = testURL
	}

	req, err := http.NewRequest(http.MethodDelete, base+"/payments/"+strconv.Itoa(id), nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(ic.APIKey, "")

	resp, err := ic.Http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	statusCode := resp.StatusCode

	if statusCode >= http.StatusBadRequest {
		return errors.New("Received " + strconv.Itoa(statusCode) + " response against request DELETE " + req.URL.String() + ": " + string(respBody))
	}

	return nil
}

func main() {
	// set up invoiced client
	client := NewInvoicedClient(InvoicedApiKey, IsInvoicedSandbox)

	begin, err := time.Parse("2006-01-02", dateRangeBegin)
	if err != nil {
		fmt.Println("error in timestamp parsing: " + err.Error())
		return
	}
	end, err := time.Parse("2006-01-02", dateRangeEnd)
	if err != nil {
		fmt.Println("error in timestamp parsing: " + err.Error())
		return
	}

	paymentsInRange, err := client.getAllPaymentsInRange(begin.Unix(), end.Unix())
	if err != nil {
		panic(err)
	}

	for _, p := range paymentsInRange {
		if p.Voided {
			fmt.Println("Skipping payment with ID #" + strconv.Itoa(p.ID) + " because it is already voided")
			continue
		}
		err := client.deletePayment(p.ID)
		if err != nil {
			fmt.Println("Error deleting payment with ID #" + strconv.Itoa(p.ID) + ": " + err.Error())
		} else {
			fmt.Println("Success: payment with ID #" + strconv.Itoa(p.ID) + " is voided")
		}
	}
}

