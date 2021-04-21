package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Export Customer With Magic Login Links
// This script prints CSV-formatted data to excel in the following format:
// Customer Name | Customer Number | Magic Link
// string        | string          | URL (string)

const (
	baseURL = "https://api.invoiced.com"
	testURL = "https://api.sandbox.invoiced.com"



	TokenTTL = time.Hour * 24 * 90 // FIXME: Update token TTL as desired; default is 90 days

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

	sandBoxEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "your environment production or sandbox")
	companyUsername := flag.String("companyusername","","your company username in Settings > Customer Portal")
	magiclinkkey := flag.String("magiclinkkey","","your magic link key in Settings > Developers")
	fileLocation := flag.String("file", "", "specify your excel file")

	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("This program will generate user registrations")

	if *key == "" {
		fmt.Print("Please enter your API Key: ")
		*key, _ = reader.ReadString('\n')
		*key = strings.TrimSpace(*key)
	}

	*environment = strings.ToUpper(strings.TrimSpace(*environment))

	if *environment == "" {
		fmt.Println("Enter P for Production, S for Sandbox: ")
		*environment, _ = reader.ReadString('\n')
		*environment = strings.TrimSpace(*environment)
	}

	if *environment == "P" {
		sandBoxEnv = false
		fmt.Println("Using Production for the environment")
	} else if *environment == "S" {
		fmt.Println("Using Sandbox for the environment")
	} else {
		fmt.Println("Unrecognized value ", *environment, ", enter P or S only")
		return
	}

	if *companyUsername == "" {
		fmt.Println("Please specify your company's username which can be found in Settings > Customer Portal: ")
		*companyUsername, _ = reader.ReadString('\n')
		*companyUsername = strings.TrimSpace(*companyUsername)
	}

	if *magiclinkkey == "" {
		fmt.Println("Please specify your magic link key can be found in Settings > Developers: ")
		*magiclinkkey, _ = reader.ReadString('\n')
		*magiclinkkey = strings.TrimSpace(*magiclinkkey)
	}

	if *fileLocation == "" {
		fmt.Println("Please specify your excel file: ")
		*fileLocation, _ = reader.ReadString('\n')
		*fileLocation = strings.TrimSpace(*fileLocation)
	}

	*fileLocation = strings.TrimSpace(*fileLocation)


	// set up invoiced client
	client := NewInvoicedClient(*key, sandBoxEnv)

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

		tokenString, err := token.SignedString([]byte(*magiclinkkey ))
		if err != nil {
			fmt.Println(err.Error())
		}

		_ = f.SetCellValue("Sheet1", "A"+strconv.Itoa(rowNum), c.Name)
		_ = f.SetCellValue("Sheet1", "B"+strconv.Itoa(rowNum), c.Number)
		_ = f.SetCellValue("Sheet1", "C"+strconv.Itoa(rowNum), c.Email)

		magiclinkURL := "https://"+*companyUsername+".invoiced.com/login/"+tokenString

		if sandBoxEnv {
			magiclinkURL = "https://"+*companyUsername+".sandbox.invoiced.com/login/"+tokenString
		}
		_ = f.SetCellValue("Sheet1", "D"+strconv.Itoa(rowNum),
			magiclinkURL)

		rowNum += 1
	}

	if err = f.SaveAs(*fileLocation); err != nil {
		panic(err)
	}
}
