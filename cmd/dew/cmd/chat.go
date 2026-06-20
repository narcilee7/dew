package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/narcilee7/dew/pkg/core"
	"github.com/narcilee7/dew/pkg/event"
	"github.com/narcilee7/dew/pkg/llm"
	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive chat session",
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

		runner := core.NewRunner(h.Provider, h.Registry)
		runner.Logger = h.Logger
		runner.Context = &core.SimpleContextManager{SystemPrompt: cfg.SystemPrompt}

		reader := bufio.NewReader(os.Stdin)
		fmt.Println("=== dew chat ===")
		for {
			fmt.Print("> ")
			line, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if line == "quit" || line == "exit" {
				break
			}

			_ = sess.Append(ctx, llm.Message{Role: core.RoleUser, Content: line})

			events := make(chan event.Event, 64)
			done := make(chan error, 1)
			go func() {
				done <- runner.Run(ctx, sess, core.RunOptions{
					MaxTurns: 1,
					Timeout:  time.Duration(cfg.TimeoutMs) * time.Millisecond,
				}, events)
				close(events)
			}()

			for ev := range events {
				printEvent(ev)
			}
			if err := <-done; err != nil {
				fmt.Printf("[error] %v\n", err)
			}
		}
		return nil
	},
}
