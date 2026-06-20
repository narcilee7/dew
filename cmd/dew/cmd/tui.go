package cmd

import (
	"context"
	"strings"
	"time"

	"github.com/narcilee7/dew/pkg/ai"
	"github.com/narcilee7/dew/pkg/core"
	"github.com/narcilee7/dew/pkg/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui [message]",
	Short: "Run dew with an interactive terminal UI",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := DefaultConfig()
		rt, err := NewRuntime(cfg)
		if err != nil {
			return err
		}

		ctx := context.Background()
		sess, err := rt.NewSession(ctx)
		if err != nil {
			return err
		}
		defer sess.Close()

		message := strings.Join(args, " ")
		if message == "" {
			message = "hello dew"
		}
		_ = sess.Append(ctx, ai.Message{Role: core.RoleUser, Content: message})

		return tui.Run(rt.Harness, sess, core.RunOptions{
			MaxTurns: cfg.MaxTurns,
			Timeout:  time.Duration(cfg.TimeoutMs) * time.Millisecond,
			Model:    cfg.Model,
		})
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
