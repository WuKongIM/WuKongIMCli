package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type LiMaoContext struct {
	opts *Options
}

func NewLiMaoContext() *LiMaoContext {
	c := &LiMaoContext{
		opts: NewOptions(),
	}
	err := c.opts.Load()
	if err != nil {
		panic(err)
	}
	return c
}

type LiMao struct {
	rootCmd *cobra.Command
}

func NewLiMao() *LiMao {

	return &LiMao{
		rootCmd: &cobra.Command{
			Use:   "lim",
			Short: "LiMaoIM 简洁，性能强劲的分布式即时通讯系统",
			Long:  `LiMaoIM 简洁，性能强劲的分布式即时通讯系统 详情查看文档：https://docs.limaoim.cn`,
			CompletionOptions: cobra.CompletionOptions{
				DisableDefaultCmd: true,
			},
		},
	}
}

func (l *LiMao) addCommand(cmd CMD) {
	l.rootCmd.AddCommand(cmd.CMD())
}

func (l *LiMao) Execute() {
	ctx := NewLiMaoContext()
	l.addCommand(newBenchCMD(ctx))
	l.addCommand(newContextCMD(ctx))
	l.addCommand(newTopCMD(ctx))

	if err := l.rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
