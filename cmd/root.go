package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type WuKongIMContext struct {
	opts *Options
}

func NewWuKongIMContext() *WuKongIMContext {
	c := &WuKongIMContext{
		opts: NewOptions(),
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
			Short: "WuKongIM 简洁，性能强劲的分布式即时通讯系统",
			Long:  `WuKongIM 简洁，性能强劲的分布式即时通讯系统 详情查看文档：https://docs.WuKongIM.cn`,
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
	ctx := NewWuKongIMContext()
	l.addCommand(newBenchCMD(ctx))
	l.addCommand(newContextCMD(ctx))
	l.addCommand(newTopCMD(ctx))

	if err := l.rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
