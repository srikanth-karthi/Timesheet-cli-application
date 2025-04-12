package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `To enable shell autocompletion:

Bash:

  $ source <(timesheet completion bash)
  $ timesheet completion bash > /etc/bash_completion.d/timesheet

Zsh:

  $ timesheet completion zsh > "${fpath[1]}/_timesheet"
  $ autoload -U compinit; compinit

Fish:

  $ timesheet completion fish | source
  $ timesheet completion fish > ~/.config/fish/completions/timesheet.fish

PowerShell:

  PS> timesheet completion powershell | Out-String | Invoke-Expression
  PS> timesheet completion powershell > timesheet.ps1
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
