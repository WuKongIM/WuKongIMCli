package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Version string
var Commit string
var CommitDate string

type WuKongIMContext struct {
	opts *Options
	w    *WuKongIM
}

func NewWuKongIMContext(w *WuKongIM) *WuKongIMContext {
	c := &WuKongIMContext{
		opts: NewOptions(),
		w:    w,
	}
	err := c.opts.Load()
	if err != nil {
		panic(err)
	}
	return c
}

type WuKongIM struct {
	rootCmd *cobra.Command
}

func NewWuKongIM() *WuKongIM {

	return &WuKongIM{
		rootCmd: &cobra.Command{
			Use:   "wk",
			Short: "WuKongIM is a concise and high-performance distributed instant messaging system",
			Long:  `This is a brief introduction to WuKongIM, a high-performance distributed instant messaging system. For more information, please refer to the documentation at https://githubim.com.`,
			CompletionOptions: cobra.CompletionOptions{
				DisableDefaultCmd: true,
			},
		},
	}
}

func (l *WuKongIM) addCommand(cmd CMD) {
	l.rootCmd.AddCommand(cmd.CMD())
}

func (l *WuKongIM) Execute() {
	ctx := NewWuKongIMContext(l)
	l.addCommand(newBenchCMD(ctx))
	l.addCommand(newContextCMD(ctx))
	l.addCommand(newTopCMD(ctx))
	l.addCommand(newStartCMD(ctx))
	l.addCommand(newDoctorCMD(ctx))
	l.addCommand(newUpgradeCMD(ctx))

	if err := l.rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
