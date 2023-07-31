package cmd

import (
	"doocli/ai"
	"doocli/db"
	"doocli/utils"
	"fmt"
	"github.com/spf13/cobra"
	"net/url"
	"os"
)

var aiCmd = &cobra.Command{
	Use:   "ai",
	Short: "Start AI service",
	PreRun: func(cmd *cobra.Command, args []string) {
		if ai.HttpPort == "" {
			utils.PrintError("httpPort is required")
			os.Exit(1)
		}
		if ai.ServerUrl != "" {
			_, err := url.Parse(ai.ServerUrl)
			if err != nil {
				utils.PrintError(fmt.Sprintf("serverUrl is invalid: %s", err.Error()))
				os.Exit(1)
			}
		}
		if ai.ClaudeAgency != "" {
			_, err := url.Parse(ai.ClaudeAgency)
			if err != nil {
				utils.PrintError(fmt.Sprintf("claudeAgency is invalid: %s", err.Error()))
				os.Exit(1)
			}
		}
		if ai.OpenaiAgency != "" {
			_, err := url.Parse(ai.OpenaiAgency)
			if err != nil {
				utils.PrintError(fmt.Sprintf("openaiAgency is invalid: %s", err.Error()))
				os.Exit(1)
			}
		}
		db.Init()
	},
	Run: func(cmd *cobra.Command, args []string) {
		ai.Start()
	},
}

func init() {
	rootCmd.AddCommand(aiCmd)
	aiCmd.Flags().StringVar(&ai.HttpPort, "httpPort", "8881", "start http service port")
	aiCmd.Flags().StringVar(&ai.ServerUrl, "serverUrl", "", "server api url")
	aiCmd.Flags().StringVar(&ai.ClaudeToken, "claudeToken", "", "claude.ai token")
	aiCmd.Flags().StringVar(&ai.ClaudeAgency, "claudeAgency", "", "claude.ai request proxy url")
	aiCmd.Flags().StringVar(&ai.OpenaiKey, "openaiKey", "", "openai api key")
	aiCmd.Flags().StringVar(&ai.OpenaiAgency, "openaiAgency", "", "openai request proxy url")
}
