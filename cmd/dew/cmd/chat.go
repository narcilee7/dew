package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/narcilee7/dew/pkg/core"
	"github.com/narcilee7/dew/pkg/llm"
	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive chat session",
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

			done := make(chan error, 1)
			go func() {
				done <- rt.Harness.Run(ctx, sess, core.RunOptions{
					MaxTurns: 1,
					Timeout:  time.Duration(cfg.TimeoutMs) * time.Millisecond,
				})
			}()

			for ev := range rt.Harness.Events() {
				printEvent(ev)
			}
			if err := <-done; err != nil {
				fmt.Printf("[error] %v\n", err)
			}
		}
		return nil
	},
}
