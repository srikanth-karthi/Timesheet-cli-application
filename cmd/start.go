package cmd

import (
	"bufio"

	"fmt"
	"github.com/spf13/cobra"
	"github.com/srikanth-karthi/timesheet/internal/setup"
	"log"
	"os"
	"strings"
	"time"
	"google.golang.org/api/sheets/v4"
	"github.com/srikanth-karthi/timesheet/internal"
)

var bucketFlag string

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start tracking time",
	Run: func(cmd *cobra.Command, args []string) {
		if !internal.IsLoggedIn() {
			fmt.Println("❌ Please run 'timesheet setup' first.")
			os.Exit(1)
		}
		provider := setup.GetCredentialProvider()
		srv := setup.GetSheetsService(provider)

		userSheet := internal.CurrentUserID
		meta, _ := internal.LoadMeta()

		if meta.SessionStart != "" {
			fmt.Printf("⚠️  A session is already running (started at %s).\n", meta.SessionStart)
			fmt.Print("❓ Do you want to abandon it and start a new session? (yes/no): ")

			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			answer = strings.ToLower(strings.TrimSpace(answer))

			if answer != "yes" && answer != "y" {
				fmt.Println("🚫 Aborting. Existing session still active.")
				os.Exit(0)
			}

			oldStartTime, err := time.Parse(time.RFC3339, meta.SessionStart)
			if err != nil {
				log.Fatalf("❌ Invalid session_start time: %v", err)
			}

			rowsResp, err := srv.Spreadsheets.Values.Get(spreadsheetID, userSheet+"!A5:G").Do()
			if err != nil {
				log.Fatalf("❌ Failed to read timesheet rows: %v", err)
			}

			rowIndex := -1
			for i, row := range rowsResp.Values {

				if len(row) >= 6 {
					tsRaw := fmt.Sprintf("%v", row[5])
					parsedSheetTime, err1 := time.Parse(time.RFC3339, tsRaw)
					parsedMetaTime, err2 := time.Parse(time.RFC3339, meta.SessionStart)
					if err1 != nil || err2 != nil {
						continue
					}
					if parsedSheetTime.Equal(parsedMetaTime) {
						rowIndex = i + 3 // sheet starts from A3
						break
					}
				}
			}

			if rowIndex != -1 {
				now := time.Now()
				duration := now.Sub(oldStartTime).Hours()
				hours := fmt.Sprintf("%.2f", duration)

				cellRef := fmt.Sprintf("E%d", rowIndex) // fixed column to E
				_, err = srv.Spreadsheets.Values.Update(spreadsheetID, userSheet+"!"+cellRef, &sheets.ValueRange{
					Values: [][]interface{}{{hours}},
				}).ValueInputOption("USER_ENTERED").Do()
				if err != nil {
					log.Fatalf("❌ Failed to update hours: %v", err)
				}
				fmt.Printf("🕒 Previous session duration: %s hrs\n", hours)
			} else {
				fmt.Println("⚠️ Could not find previous session row to log hours.")
			}

			meta.SessionStart = ""
			_ = internal.SaveMeta(meta)
			fmt.Println("🗑️  Previous session ended and logged.")
		}

		bucket := bucketFlag
		if bucket == "" {
			bucket = meta.Active
			if bucket == "" {
				log.Fatalf("❌ No active bucket found. Use 'timesheet bucket <name>' to set one.")
			}
		}

		bucketResp, err := srv.Spreadsheets.Values.Get(spreadsheetID, userSheet+"!C1:Z1").Do()
		if err != nil {
			log.Fatalf("❌ Could not fetch buckets: %v", err)
		}

		valid := false
		if len(bucketResp.Values) > 0 {
			for _, cell := range bucketResp.Values[0] {
				if cell.(string) == bucket {
					valid = true
					break
				}
			}
		}
		if !valid {
			log.Fatalf("❌ Bucket '%s' is not valid. Use 'timesheet bucket' to view available ones.", bucket)
		}

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("📝 Task description: ")
		desc, _ := reader.ReadString('\n')
		desc = strings.TrimSpace(desc)

		startTime := time.Now()
		startTimeRFC := startTime.Format(time.RFC3339)
		formattedDate := startTime.Format("02/01/06") // dd/mm/yy
		day := startTime.Format("Monday")             // Weekday

		meta.SessionStart = startTimeRFC
		_ = internal.SaveMeta(meta)

		_, err = srv.Spreadsheets.Values.Append(spreadsheetID, userSheet+"!A3:G", &sheets.ValueRange{
			Values: [][]interface{}{
				{formattedDate, day, bucket, desc, "", startTimeRFC},
			},
		}).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()

		if err != nil {
			log.Fatalf("❌ Failed to log new task: %v", err)
		}

		fmt.Printf("⏱️  Started tracking task: '%s' in bucket '%s'\n", desc, bucket)
	},
}
