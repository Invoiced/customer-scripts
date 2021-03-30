package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/Invoiced/invoiced-go/invdendpoint"
	"github.com/invoiced/invoiced-go"
	"os"
	"strconv"
	"strings"
)

//This program will import in customer contacts

const sheet = "Sheet1"

func FindRoleId(roles invdapi.Roles, roleName string) (roleId string, err error) {

	for _, role := range roles {
		if strings.ToLower(strings.TrimSpace(role.Name)) == roleName {
			return role.Id, nil
		}
	}


	return "", errors.New("Could not locate role with name = " + roleName)

}


func main() {
	sandBoxEnv := true
	key := flag.String("key", "", "api key in Settings > Developer")
	environment := flag.String("env", "", "your environment production or sandbox")
	fileLocation := flag.String("file", "", "specify your excel file")

	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("This program will upload notifications")

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

	if *fileLocation == "" {
		fmt.Println("Please specify your excel file: ")
		*fileLocation, _ = reader.ReadString('\n')
		*fileLocation = strings.TrimSpace(*fileLocation)
	}

	*fileLocation = strings.TrimSpace(*fileLocation)


	//Load notifications and roles into memory

	//Load all roles from Invoices

	conn := invdapi.NewConnection(*key, sandBoxEnv)

	invoicedRoles, err := conn.NewRole().ListAll(nil,nil)

	if err != nil {
		fmt.Println("Ran into error fetching roles from Invoiced, err => ",err)
		return
	}


	// Notification role map
	notificationRoleMap := make(map[string][]string)

	f, err := excelize.OpenFile(*fileLocation)

	if err != nil {
		fmt.Println("Ran into error opening file ",*fileLocation, "error => ",err)
		return
	}

	fmt.Println("Read in excel file ", fileLocation, ", successfully")

	roleIndex := "A"
	notificationRoleIndex := "B"

	rows, err := f.GetRows(sheet)

	if err != nil {
		panic("Error trying to get rows for the sheet" + err.Error())
	}

	for i, row := range rows {
		if i == 0 {
			fmt.Println("Skipping header row")
			fmt.Println(row)
			continue
		}

		role, _ := f.GetCellValue(sheet,roleIndex + strconv.Itoa(i + 1))
		notificationEventId,_ :=  f.GetCellValue(sheet,notificationRoleIndex + strconv.Itoa(i + 1))

		role = strings.ToLower(strings.TrimSpace(role))
		roleID, err := FindRoleId(invoicedRoles,role)

		if err != nil {
			fmt.Println("Could not find role in Invoiced ", role)
			continue
		}

		notificationEventId = strings.TrimSpace(notificationEventId)

		_, found := notificationRoleMap[roleID]

		if !found {
			notificationRoleMap[roleID] = make([]string,0)
		}

		fmt.Println("Adding role id = ",roleID, "for notification event ", notificationEventId)
		notificationRoleMap[roleID] = append(notificationRoleMap[roleID],notificationEventId)

	}

	//Load a list of all users

	userConn := conn.NewUser()

	users, err := userConn.ListAll(nil,nil)

	if err != nil {
		fmt.Println("Got an error fetching users => ",err)
		return
	} else {
		fmt.Println("Fetched",len(users), "users")
	}

	//delete all notifications for all the users
	fmt.Println("Deleting notifications for all users")

	notifications, err := conn.NewNotification().ListAll(nil,nil)

	if err != nil {
		fmt.Println("Got an error deleting notifications => ",err)
		return
	}

	for _, notification := range notifications {
		fmt.Println("Deleting notification for ",notification.UserId,notification.Event)
		err = notification.Delete(notification.Id)
		if err != nil {
			fmt.Println("Got an error deleting notification ", notification.Id, ", err => ", err)
		} else {
			fmt.Println("Successfully deleted notification ",notification.Id)
		}
	}

	for _, user := range users {

		userRole := user.Role
		userId := user.User.Id
		userEmail := user.User.Email

		userNotificationEvents, found := notificationRoleMap[userRole]

		if !found {
			fmt.Println("Did not find matching user role ", userRole," in excel sheet", "for user ", userEmail)
			continue
		}

		fmt.Println("Creating notifications for user ", userEmail)

		for _, userNotificationEvent := range userNotificationEvents {
			if userNotificationEvent == "never" || userNotificationEvent == "day" || userNotificationEvent == "month" || userNotificationEvent == "week" {

				if userNotificationEvent == "never" {
					userNotificationEvent = ""
				}

				_,err = user.SetUserEmailFrequency(userNotificationEvent,user.Id)

				if err != nil {
					fmt.Println("Error updating frequency to ",userNotificationEvent," for user = ", user.User.Email)
				}

				continue

			}

			notificationRequest := new(invdendpoint.NotificationRequest)
			notificationRequest.Enabled = true
			notificationRequest.Event = userNotificationEvent
			notificationRequest.Medium = "email"
			notificationRequest.UserId = userId

			notifResponse, err := conn.NewNotification().Create(notificationRequest)

			if err != nil {
				fmt.Println("Could not add notification ", userNotificationEvent, " for user ", userEmail)
				continue
			}

			fmt.Println("Successfully created notification ", userNotificationEvent, "for user", userEmail, " with id = ",notifResponse.Id)

		}


	}

}

