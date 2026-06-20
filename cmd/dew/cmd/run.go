package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/narcilee7/dew/pkg/core"
	"github.com/narcilee7/dew/pkg/event"
	"github.com/narcilee7/dew/pkg/llm"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [prompt]",
	Short: "Run dew with a single prompt",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := DefaultConfig()
		h, err := NewHarness(cfg)
		if err != nil {
			return err
		}

		ctx := context.Background()
		sess, err := h.NewSession(ctx)
		if err != nil {
			return err
		}
		defer sess.Close()

		_ = sess.Append(ctx, llm.Message{
			Role:    core.RoleUser,
			Content: args[0],
		})

		runner := core.NewRunner(h.Provider, h.Registry)
		runner.Logger = h.Logger
		runner.Context = &core.SimpleContextManager{SystemPrompt: cfg.SystemPrompt}

		events := make(chan event.Event, 64)
		done := make(chan error, 1)
		go func() {
			done <- runner.Run(ctx, sess, core.RunOptions{
				MaxTurns: cfg.MaxTurns,
				Timeout:  time.Duration(cfg.TimeoutMs) * time.Millisecond,
			}, events)
			close(events)
		}()

		for ev := range events {
			printEvent(ev)
		}

		if err := <-done; err != nil {
			return err
		}
		return nil
	},
}

func printEvent(ev event.Event) {
	switch e := ev.(type) {
	case event.AgentStartEvent:
		fmt.Printf("[start] agent=%s\n", e.AgentID)
	case event.TurnStartEvent:
		fmt.Printf("[turn] %d\n", e.Turn)
	case event.TextDeltaEvent:
		fmt.Printf("[text] %s\n", e.Delta)
	case event.ToolCallStartEvent:
		fmt.Printf("[tool-start] %s(%s)\n", e.Call.Name, e.Call.ID)
	case event.ToolResultEvent:
		for _, part := range e.Result.Content {
			fmt.Printf("[tool-result] %s: %s\n", e.Name, part.Text)
		}
		if e.Result.Metadata.Error != "" {
			fmt.Printf("[tool-error] %s: %s\n", e.Name, e.Result.Metadata.Error)
		}
	case event.TurnEndEvent:
		fmt.Printf("[/turn] %d\n", e.Turn)
	case event.AgentEndEvent:
		fmt.Printf("[end] agent=%s\n", e.AgentID)
	case event.ErrorEvent:
		fmt.Printf("[error] %v\n", e.Err)
	}
}
