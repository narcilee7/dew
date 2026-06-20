package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage sessions",
}

var sessionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List active sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Active sessions:")
		fmt.Println("(session persistence not yet implemented)")
		return nil
	},
}

var sessionResumeCmd = &cobra.Command{
	Use:   "resume [session-id]",
	Short: "Resume a session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Resuming session %s (not yet implemented)\n", args[0])
		return nil
	},
}

func init() {
	sessionCmd.AddCommand(sessionListCmd)
	sessionCmd.AddCommand(sessionResumeCmd)
}
