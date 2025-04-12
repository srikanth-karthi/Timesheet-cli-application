package cmd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/srikanth-karthi/timesheet/internal"
	"github.com/srikanth-karthi/timesheet/internal/setup"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var spreadsheetID = "1VWNJK55ytijKrR8QcOEr1JORUSWtlcoudkkTbpsUiYM"
var createUser bool

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Authenticate and set up your timesheet",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔐 Starting timesheet setup...")

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter your EMP ID: ")
		empID, _ := reader.ReadString('\n')
		empID = strings.TrimSpace(empID)

		fmt.Print("Enter your password: ")
		password, _ := reader.ReadString('\n')
		password = strings.TrimSpace(password)

		ctx := context.Background()
		provider := setup.GetCredentialProvider()

		creds, err := provider.GetJSON()
		if err != nil {
			log.Fatalf("  Failed to load credentials: %v", err)
		}

		srv, err := sheets.NewService(ctx, option.WithCredentialsJSON(creds))
		if err != nil {
			log.Fatalf("  Unable to create Sheets service: %v", err)
		}

		if err != nil {
			log.Fatalf("  Unable to create Sheets service: %v", err)
		}

		err = ensureAdminSheet(srv, spreadsheetID)
		if err != nil {
			log.Fatalf("  Failed to ensure admin sheet: %v", err)
		}

		if createUser {
			err := createUserInAdmin(srv, spreadsheetID, empID, password)
			if err != nil {
				log.Fatalf("  Failed to create user: %v", err)
			}
			log.Printf("   User %s created successfully", empID)
			log.Printf("Would you like to provide a project folder path? (y/n)")
		
			reader := bufio.NewReader(os.Stdin)
			answer, err := reader.ReadString('\n')
			if err != nil {
				log.Fatalf("  Failed to read input: %v", err)
			}
			answer = strings.ToLower(strings.TrimSpace(answer))
		
			if answer != "yes" && answer != "y" {
				fmt.Println("🚫 Aborting. Existing session still active.")
				os.Exit(0)
			}
	
		
		} else {
			ok, err := validateCredentials(srv, spreadsheetID, empID, password)
			if err != nil {
				log.Fatalf("  Failed to validate login: %v", err)
			}
			if !ok {
				log.Fatalf("  Invalid EMP ID or password")
			}
			log.Printf("   Welcome, %s!", empID)
		}
		
	},
}

func ensureAdminSheet(srv *sheets.Service, spreadsheetID string) error {
	ss, err := srv.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return err
	}

	exists := false
	for _, sheet := range ss.Sheets {
		if sheet.Properties.Title == "admin" {
			exists = true
			break
		}
	}

	if !exists {
		_, err := srv.Spreadsheets.BatchUpdate(spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{
				{
					AddSheet: &sheets.AddSheetRequest{
						Properties: &sheets.SheetProperties{
							Title: "admin",
						},
					},
				},
			},
		}).Do()
		if err != nil {
			return err
		}

		_, err = srv.Spreadsheets.Values.Update(spreadsheetID, "admin!A1:B1", &sheets.ValueRange{
			Values: [][]interface{}{
				{"emp_id", "password"},
			},
		}).ValueInputOption("RAW").Do()
		if err != nil {
			return err
		}

	} else {

	}

	return nil
}

func validateCredentials(srv *sheets.Service, spreadsheetID, empID, password string) (bool, error) {
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, "admin!A2:B").Do()
	if err != nil {
		return false, err
	}

	for _, row := range resp.Values {
		if len(row) < 2 {
			continue
		}
		sheetEmpID := fmt.Sprintf("%v", row[0])
		sheetPass := fmt.Sprintf("%v", row[1])

		if empID == sheetEmpID && password == sheetPass {
			err = internal.SaveSession(empID)
			if err != nil {
				log.Fatalf("  Failed to save session: %v", err)
			}

			return true, nil
		}
	}

	return false, nil
}

func createUserInAdmin(srv *sheets.Service, spreadsheetID, empID, password string) error {
	exists, _ := validateCredentials(srv, spreadsheetID, empID, password)
	if exists {
		return fmt.Errorf("user already exists")
	}

	_, err := srv.Spreadsheets.Values.Append(spreadsheetID, "admin!A2:B",
		&sheets.ValueRange{
			Values: [][]interface{}{{empID, password}},
		},
	).ValueInputOption("RAW").Do()
	if err != nil {
		return err
	}

	userSheetName := empID
	if err := ensureUserSheet(srv, spreadsheetID, userSheetName); err != nil {
		return fmt.Errorf("failed to create user sheet: %v", err)
	}
	err = internal.SaveSession(empID)
	if err != nil {
		log.Fatalf("  Failed to save session: %v", err)
	}

	return nil
}

func ensureUserSheet(srv *sheets.Service, spreadsheetID, sheetName string) error {
	ss, err := srv.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return err
	}

	for _, sheet := range ss.Sheets {
		if sheet.Properties.Title == sheetName {
			log.Printf("ℹ️ Sheet '%s' already exists.", sheetName)
			return nil
		}
	}

	_, err = srv.Spreadsheets.BatchUpdate(spreadsheetID, &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{
			{
				AddSheet: &sheets.AddSheetRequest{
					Properties: &sheets.SheetProperties{
						Title: sheetName,
					},
				},
			},
		},
	}).Do()
	if err != nil {
		return err
	}
	_, err = srv.Spreadsheets.Values.Update(spreadsheetID, sheetName+"!A1:G2", &sheets.ValueRange{
		Values: [][]any{
			{"meta", "buckets", "general"},
			{"date", "day", "project", "task_description", "hours", "timestamp"},
		},
	}).ValueInputOption("USER_ENTERED").Do()

	if err != nil {
		return err
	}

	log.Printf("   Created sheet '%s' with headers + meta.", sheetName)
	return nil
}
