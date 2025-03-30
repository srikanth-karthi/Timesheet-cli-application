package cmd

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"
"os/exec"
"path/filepath"
	"github.com/spf13/cobra"
	"github.com/srikanth-karthi/timesheet/internal"
	"github.com/srikanth-karthi/timesheet/internal/setup"
)

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

		// 📅 Get current week (Mon–Sun)
		now := time.Now()
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday = 7
		}
		monday := now.AddDate(0, 0, -weekday+1)
		sunday := monday.AddDate(0, 0, 6)

		// 📊 Fetch all data
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
			if len(row) < 6 {
				continue
			}
			dateStr := fmt.Sprintf("%v", row[0])
			project := fmt.Sprintf("%v", row[2])
			description := fmt.Sprintf("%v", row[3])
			hoursStr := fmt.Sprintf("%v", row[4])

			date, err := time.Parse("02/01/06", dateStr)
			if err != nil || date.Before(monday) || date.After(sunday) {
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

		// 🗂️ Group by day → project
		daily := map[string][]Entry{}
		projectTotals := map[string]float64{}
		total := 0.0

		for _, e := range entries {
			key := e.Date.Format("Mon (Jan 02)")
			daily[key] = append(daily[key], e)
			projectTotals[e.Project] += e.Hours
			total += e.Hours
		}

		// 📅 Print header
		_, week := monday.ISOWeek()
		fmt.Printf("\n📅 Week %d (%s–%s)\n", week, monday.Format("Jan 02"), sunday.Format("Jan 02"))
		fmt.Println(strings.Repeat("-", 26))

		// 📆 Daily breakdown
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

		// 📁 Project Totals
		fmt.Println(strings.Repeat("-", 26))
		fmt.Println("📁 Project Totals:")
		for project, hrs := range projectTotals {
			fmt.Printf("- %-10s → %.1f hrs\n", project, hrs)
		}

		fmt.Printf("\n🕒 Total Hours This Week: %.1f   \n", total)

		// 🧠 Check for Git activity in project path
		meta, err := internal.LoadMeta()
		if err != nil || meta.ProjectPath == "" {
			log.Printf("  ⚠️  No project path found or failed to load metadata.")
		} else {
			gitPath := meta.ProjectPath
		
			// 🔍 Find the first nested Git repo
			var gitRoot string
			err = filepath.Walk(gitPath, func(path string, info os.FileInfo, err error) error {
				if info == nil {
					return nil
				}
				if info.IsDir() && info.Name() == ".git" {
					gitRoot = filepath.Dir(path)
					return filepath.SkipDir
				}
				return nil
			})
		
			if gitRoot == "" {
				log.Printf("  ⚠️  No Git repo found under project path: %s", gitPath)
			} else {
				// 🔢 Get latest 10 commits by author
				cmd := exec.Command("git", "-C", gitRoot, "log", "--author=srikanthk.c", "--pretty=format:%H", "-n", "10")
				out, err := cmd.Output()
				if err != nil || len(out) == 0 {
					log.Printf("  ⚠️  No commits found by srikanth-karthi in %s", gitRoot)
					return
				}
				commits := strings.Split(strings.TrimSpace(string(out)), "\n")
		
				for _, hash := range commits {
					// 📝 Get list of files in the commit
					cmdFiles := exec.Command("git", "-C", gitRoot, "show", "--pretty=", "--name-only", hash)
					filesOut, _ := cmdFiles.Output()
					files := strings.Split(strings.TrimSpace(string(filesOut)), "\n")
		
					// 🚫 Filter out ignored files
					var hasNonIgnored bool 
					for _, file := range files {
						if file == "" {
							continue
						}
						checkCmd := exec.Command("git", "-C", gitRoot, "check-ignore", file)
						err := checkCmd.Run()
						if err != nil { // Not ignored
							hasNonIgnored = true
							break
						}
					}
		
					if hasNonIgnored {
						// 🎯 Get commit summary
						cmdSummary := exec.Command("git", "-C", gitRoot, "show", "-s", "--format=%h %ad %s", "--date=short", hash)
						summary, _ := cmdSummary.Output()
		
						fmt.Println(strings.Repeat("-", 26))
						fmt.Println("📦 Latest Non-Ignored Commit by srikanth-karthi:")
						fmt.Printf("  %s\n", string(summary))
						break
					}
				}
			}
		}
		

	},
}
