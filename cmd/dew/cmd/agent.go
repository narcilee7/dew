package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage agents",
}

var agentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Registered agents:")
		fmt.Println("(agent registry not yet implemented)")
		return nil
	},
}

func init() {
	agentCmd.AddCommand(agentListCmd)
}
