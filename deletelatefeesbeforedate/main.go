package main

import (
	"fmt"
	"github.com/Invoiced/invoiced-go"
	"github.com/Invoiced/invoiced-go/invdendpoint"
)

func main() {
	//CHANGE ME
	invdConnection := invdapi.NewConnection("API_KEY", false)
	invdInvoice := invdConnection.NewInvoice()
	filter := invdendpoint.NewFilter()
	filter.Set("closed", "false")

	invdRetInvoices, err := invdInvoice.ListAll(filter, nil)

	if err != nil {
		fmt.Println("Got error fetching invoice => ", err)
	}

	fmt.Println("Count of invoices fetched => ", len(invdRetInvoices))

	sum := 0
	for i, invoice := range invdRetInvoices {
		fmt.Println("Iteration i=", i)

		if invoice.Draft {
			fmt.Println("skipping invoice #", invoice.Number, "invoice is a draft")
			continue
		}

		//Change Me - change to a unix timestamp where all invoices dated before this timestamp will have late fees removed
		removeLateFeesBeforeUnixDate := int64(1609477200)

		if invoice.Date < removeLateFeesBeforeUnixDate {
			fmt.Println("Removing late fee for invoice number ", invoice.Number, ", invoice.date ", invoice.Date)
			sum += 1
			lateFeePresent := false
			lineItems := invoice.Items

			lineItemWithOutLateFees := make([]invdendpoint.LineItem, 0)

			for _, item := range lineItems {
				if item.Type != "late_fee" {
					lineItemWithOutLateFees = append(lineItemWithOutLateFees, item)
				} else {
					lateFeePresent = true
				}
			}

			if lateFeePresent {
				invdInvoiceToUpdate := invdConnection.NewInvoice()
				invdInvoiceToUpdate.Id = invoice.Id
				invdInvoiceToUpdate.Items = lineItemWithOutLateFees

				err := invdInvoiceToUpdate.Save()

				if err != nil {
					fmt.Println("Error updating invoice number ", invoice.Number, ",error => ", err)
				} else {
					fmt.Println("Successfully removed late fees for Invoice ")
				}
			} else {
				fmt.Println("No late fees for Invoice ", invoice.Number)
			}

			//if sum == 10 {
			//	break
			//}

		} else {
			continue
		}

	}

	fmt.Println("Total sum => ", sum)

}
