package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for silo.

To load completions:

Bash:
  $ source <(silo completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ silo completion bash > /etc/bash_completion.d/silo
  # macOS:
  $ silo completion bash > $(brew --prefix)/etc/bash_completion.d/silo

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  # To load completions for each session, execute once:
  $ silo completion zsh > "${fpath[1]}/_silo"

Fish:
  $ silo completion fish | source
  # To load completions for each session, execute once:
  $ silo completion fish > ~/.config/fish/completions/silo.fish

PowerShell:
  PS> silo completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> silo completion powershell > silo.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			_ = rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			_ = rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			_ = rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			_ = rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
