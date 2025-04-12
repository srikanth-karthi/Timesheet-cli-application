package cmd

import (
	"context"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/srikanth-karthi/timesheet/internal"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var rootCmd = &cobra.Command{
	Use:   "timesheet",
	Short: "CLI timesheet tracker for logging and managing work logs",
	Long:  `Track, manage, and report timesheets directly from the terminal.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {


	rootCmd.PersistentFlags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.AddCommand(
		setupCmd,
		startCmd,
		stopCmd,
		logCmd,
		reportCmd,
		bucketCmd,
	)

	bucketCmd.AddCommand(bucketNewCmd, bucketListCmd)

	setupCmd.Flags().BoolVar(&createUser, "create", false, "Create a new user during setup")
	startCmd.Flags().StringVar(&bucketFlag, "bucket", "", "Bucket to log task in")
	logCmd.Flags().StringVar(&logTask, "task", "", "Task description (required)")
	logCmd.Flags().StringVar(&logHours, "hours", "", "Hours spent (required)")
	logCmd.Flags().StringVar(&logBucket, "bucket", "", "Bucket/project name (optional)")
	logCmd.Flags().StringVar(&logDate, "date", "", "Date in dd/mm/yy format (optional)")
	reportCmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show all entries instead of just this week")

	bucketCmd.ValidArgsFunction = completeBuckets
}

// ✨ Shell completion for bucket names
func completeBuckets(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if !internal.IsLoggedIn() {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	userSheet := internal.CurrentUserID

	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("credentials.json"))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, userSheet+"!C1:Z1").Do()
	if err != nil || len(resp.Values) == 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var suggestions []string
	for _, cell := range resp.Values[0] {
		if str, ok := cell.(string); ok && strings.HasPrefix(str, toComplete) {
			suggestions = append(suggestions, str)
		}
	}

	return suggestions, cobra.ShellCompDirectiveNoFileComp
}
