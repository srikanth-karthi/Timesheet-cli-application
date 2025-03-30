package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/srikanth-karthi/timesheet/internal"
	"github.com/srikanth-karthi/timesheet/internal/setup"
	"google.golang.org/api/sheets/v4"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "⏹️ Stop tracking the current session and log the duration",
	Run: func(cmd *cobra.Command, args []string) {
		if !internal.IsLoggedIn() {
			fmt.Println("  Please run 'timesheet setup' first.")
			os.Exit(1)
		}

		meta, _ := internal.LoadMeta()
		if meta.SessionStart == "" {
			fmt.Println("⚠️ No session is currently running.")
			return
		}

		//    Parse the session start time
		oldStartTime, err := time.Parse(time.RFC3339, meta.SessionStart)
		if err != nil {
			log.Fatalf("  Invalid session_start time: %v", err)
		}

		provider := setup.GetCredentialProvider()
		srv := setup.GetSheetsService(provider)

		userSheet := internal.CurrentUserID

		//    Fetch rows from A5:G to find the matching timestamp
		rowsResp, err := srv.Spreadsheets.Values.Get(spreadsheetID, userSheet+"!A5:G").Do()
		if err != nil {
			log.Fatalf("  Failed to read timesheet rows: %v", err)
		}

		rowIndex := -1
		for i, row := range rowsResp.Values {
			if len(row) >= 6 {
				ts := fmt.Sprintf("%v", row[5])
				if ts == meta.SessionStart {
					rowIndex = i + 5 // adjust for A5:G offset
					break
				}
			}
		}

		if rowIndex != -1 {
			now := time.Now()
			duration := now.Sub(oldStartTime).Hours()
			hours := fmt.Sprintf("%.2f", duration)

			cellRef := fmt.Sprintf("E%d", rowIndex) // hours column
			_, err = srv.Spreadsheets.Values.Update(spreadsheetID, userSheet+"!"+cellRef, &sheets.ValueRange{
				Values: [][]interface{}{{hours}},
			}).ValueInputOption("USER_ENTERED").Do()

			if err != nil {
				log.Fatalf("  Failed to update hours: %v", err)
			}
			fmt.Printf("🕒 Session stopped. Duration: %s hrs logged in row %d\n", hours, rowIndex)
		} else {
			fmt.Println("⚠️ Could not find previous session row to log hours.")
		}

		//    Clear session
		meta.SessionStart = ""
		_ = internal.SaveMeta(meta)
		fmt.Println("   Session cleared.")
	},
}
