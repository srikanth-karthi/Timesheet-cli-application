package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/api/sheets/v4"

	"github.com/srikanth-karthi/timesheet/internal"
	"github.com/srikanth-karthi/timesheet/internal/setup"
)

var (
	logTask   string
	logHours  string
	logBucket string
	logDate   string
)

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "📝 Manually log a task with hours",
	Run: func(cmd *cobra.Command, args []string) {
		if !internal.IsLoggedIn() {
			fmt.Println("  Please run 'timesheet setup' first.")
			os.Exit(1)
		}

		if logTask == "" || logHours == "" {
			fmt.Println("  Please provide both --task and --hours.")
			cmd.Usage()
			os.Exit(1)
		}

		provider := setup.GetCredentialProvider()
		srv := setup.GetSheetsService(provider)

		userSheet := internal.CurrentUserID
		meta, _ := internal.LoadMeta()

		t := time.Now()
		if logDate != "" {
			parsed, err := time.Parse("02/01/06", logDate)
			if err != nil {
				log.Fatalf("  Invalid date format. Use dd/mm/yy")
			}
			t = parsed
		}
		formattedDate := t.Format("02/01/06")
		day := t.Format("Monday")

		bucket := logBucket
		if bucket == "" {
			bucket = meta.Active
			if bucket == "" {
				log.Fatalf("  No active bucket found. Use 'timesheet bucket <name>' to set one.")
			}
		}

		bucketResp, err := srv.Spreadsheets.Values.Get(spreadsheetID, userSheet+"!C1:Z1").Do()
		if err != nil {
			log.Fatalf("  Could not fetch buckets: %v", err)
		}

		valid := false
		if len(bucketResp.Values) > 0 {
			for _, cell := range bucketResp.Values[0] {
				if cellStr, ok := cell.(string); ok && cellStr == bucket {
					valid = true
					break
				}
			}
		}
		if !valid {
			log.Fatalf("  Bucket '%s' is not valid. Use 'timesheet bucket' to view available ones.", bucket)
		}

		timestamp := time.Now().Format(time.RFC3339)
		_, err = srv.Spreadsheets.Values.Append(spreadsheetID, userSheet+"!A3:G", &sheets.ValueRange{
			Values: [][]interface{}{
				{formattedDate, day, bucket, logTask, logHours, timestamp},
			},
		}).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()

		if err != nil {
			log.Fatalf("  Failed to log manual entry: %v", err)
		}

		fmt.Printf("   Logged task '%s' for %s hrs on %s [%s]\n", logTask, logHours, formattedDate, bucket)
	},
}
