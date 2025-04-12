package cmd

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/srikanth-karthi/timesheet/internal"
	"github.com/srikanth-karthi/timesheet/internal/setup"
)

var showAll bool

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "📊 Show this week's summary grouped by project",
	Run: func(cmd *cobra.Command, args []string) {
		if !internal.IsLoggedIn() {
			fmt.Println("  Please run 'timesheet setup' first.")
			os.Exit(1)
		}

		provider := setup.GetCredentialProvider()
		srv := setup.GetSheetsService(provider)
		userSheet := internal.CurrentUserID

		now := time.Now()
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday → 7
		}
		monday := now.AddDate(0, 0, -weekday+1)
		sunday := monday.AddDate(0, 0, 6)

		resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, userSheet+"!A5:G").Do()
		if err != nil {
			log.Fatalf("  Failed to fetch timesheet data: %v", err)
		}

		type Entry struct {
			Date        time.Time
			Project     string
			Description string
			Hours       float64
		}

		var entries []Entry
		for _, row := range resp.Values {
			if len(row) < 5 {
				continue
			}

			dateStr := fmt.Sprintf("%v", row[0])
			project := fmt.Sprintf("%v", row[2])
			description := fmt.Sprintf("%v", row[3])
			hoursStr := fmt.Sprintf("%v", row[4])

			date, err := time.Parse("02/01/06", dateStr)
			if err != nil {
				continue
			}

			if !showAll && (date.Before(monday) || date.After(sunday)) {
				continue
			}

			var hrs float64
			fmt.Sscanf(hoursStr, "%f", &hrs)

			entries = append(entries, Entry{
				Date:        date,
				Project:     project,
				Description: description,
				Hours:       hrs,
			})
		}

		daily := map[string][]Entry{}
		projectTotals := map[string]float64{}
		total := 0.0

		for _, e := range entries {
			key := e.Date.Format("Mon (Jan 02)")
			daily[key] = append(daily[key], e)
			projectTotals[e.Project] += e.Hours
			total += e.Hours
		}

		if showAll {
			fmt.Println("\n📊 Showing *all* timesheet entries")
		} else {
			_, week := monday.ISOWeek()
			fmt.Printf("\n📅 Week %d (%s – %s)\n", week, monday.Format("Jan 02"), sunday.Format("Jan 02"))
		}
		fmt.Println(strings.Repeat("-", 30))

		fmt.Println("\n📆 Daily Breakdown:")
		keys := make([]string, 0, len(daily))
		for k := range daily {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			ti, _ := time.Parse("Mon (Jan 02)", keys[i])
			tj, _ := time.Parse("Mon (Jan 02)", keys[j])
			return ti.Before(tj)
		})

		for _, k := range keys {
			fmt.Printf("%s\n", k)
			for _, e := range daily[k] {
				fmt.Printf("  - %-10s → %-30s → %.1f hrs\n", e.Project, e.Description, e.Hours)
			}
			fmt.Println()
		}

		fmt.Println(strings.Repeat("-", 30))
		fmt.Println("📁 Project Totals:")
		for project, hrs := range projectTotals {
			fmt.Printf("- %-10s → %.1f hrs\n", project, hrs)
		}

		fmt.Printf("\n🕒 Total Hours: %.1f\n\n", total)
	},
}


