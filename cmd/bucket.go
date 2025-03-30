package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/api/sheets/v4"

	"github.com/srikanth-karthi/timesheet/internal"
	"github.com/srikanth-karthi/timesheet/internal/setup"
)

var bucketCmd = &cobra.Command{
	Use:   "bucket [name]",
	Short: "List or switch buckets",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !internal.IsLoggedIn() {
			fmt.Println("  Please run 'timesheet setup' first.")
			os.Exit(1)
		}

		srv := getSheetsService()
		userSheet := internal.CurrentUserID

		bucketResp, err := srv.Spreadsheets.Values.Get(spreadsheetID, userSheet+"!C1:Z1").Do()
		if err != nil || len(bucketResp.Values) == 0 {
			log.Fatalf("  Could not read bucket list.")
		}
		buckets := bucketResp.Values[0]

		meta, _ := internal.LoadMeta()
		active := meta.Active

		if len(args) == 1 {
			target := args[0]
			found := false
			for _, b := range buckets {
				if b.(string) == target {
					found = true
					break
				}
			}

			if !found {
				fmt.Printf("  Bucket '%s' not found.\n", target)
				os.Exit(1)
			}

			meta.Active = target
			_ = internal.SaveMeta(meta)
			log.Printf("   Switched to bucket: %s", target)
			return
		}

		for _, cell := range buckets {
			name := cell.(string)
			prefix := "  "
			colorStart, colorEnd := "", ""
			if name == active {
				prefix = "* "
				colorStart = "\033[36m"
				colorEnd = "\033[0m"
			}
			fmt.Printf("%s%s%s%s\n", prefix, colorStart, name, colorEnd)
		}
	},
}

var bucketListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all buckets (shows current)",
	Run: func(cmd *cobra.Command, args []string) {
		if !internal.IsLoggedIn() {
			fmt.Println("  Please run 'timesheet setup' first.")
			os.Exit(1)
		}

		srv := getSheetsService()
		userSheet := internal.CurrentUserID

		bucketResp, err := srv.Spreadsheets.Values.Get(spreadsheetID, userSheet+"!C1:Z1").Do()
		if err != nil {
			log.Fatalf("  Failed to fetch buckets: %v", err)
		}

		meta, _ := internal.LoadMeta()
		active := meta.Active

		if len(bucketResp.Values) == 0 {
			fmt.Println("ℹ️ No buckets found.")
			return
		}

		for _, cell := range bucketResp.Values[0] {
			name := cell.(string)
			prefix := "  "
			colorStart, colorEnd := "", ""
			if name == active {
				prefix = "* "
				colorStart = "\033[36m"
				colorEnd = "\033[0m"
			}
			fmt.Printf("%s%s%s%s\n", prefix, colorStart, name, colorEnd)
		}
	},
}

var bucketNewCmd = &cobra.Command{
	Use:   "new [name]",
	Short: "Create or switch to a bucket",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !internal.IsLoggedIn() {
			fmt.Println("  Please run 'timesheet setup' first.")
			os.Exit(1)
		}

		bucket := args[0]
		userSheet := internal.CurrentUserID
		srv := getSheetsService()

		resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, userSheet+"!C1:Z1").Do()
		if err != nil {
			log.Fatalf("  Failed to read meta-buckets row: %v", err)
		}

		row := resp.Values
		found := false
		numBuckets := 0

		if len(row) > 0 {
			for _, cell := range row[0] {
				if str, ok := cell.(string); ok {
					numBuckets++
					if str == bucket {
						found = true
						break
					}
				}
			}
		}

		if !found {
			newCol := 3 + numBuckets
			colLetter := columnNumberToLetter(newCol)
			cellRef := fmt.Sprintf("%s1", colLetter)

			_, err = srv.Spreadsheets.Values.Update(spreadsheetID, userSheet+"!"+cellRef, &sheets.ValueRange{
				Values: [][]interface{}{{bucket}},
			}).ValueInputOption("RAW").Do()
			if err != nil {
				log.Fatalf("  Failed to append new bucket: %v", err)
			}
			log.Printf("🌟 Created new bucket: %s", bucket)
		}

		meta, _ := internal.LoadMeta()
		meta.Active = bucket
		_ = internal.SaveMeta(meta)

		log.Printf("   Switched to bucket: %s", bucket)
	},
}

func columnNumberToLetter(n int) string {
	letters := ""
	for n > 0 {
		n--
		letters = string('A'+(n%26)) + letters
		n /= 26
	}
	return letters
}

func getSheetsService() *sheets.Service {
	provider := setup.GetCredentialProvider()
	return setup.GetSheetsService(provider)
}
