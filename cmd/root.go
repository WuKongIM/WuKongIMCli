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
	l.addCommand(newContextCMD(ctx))    // 上下文命令
	l.addCommand(newBenchCMD(ctx))      // 压力测试命令
	l.addCommand(newTopCMD(ctx))        // top命令
	l.addCommand(newStartCMD(ctx))      // 启动命令
	l.addCommand(newDoctorCMD(ctx))     // 检查命令
	l.addCommand(newUpgradeCMD(ctx))    // 升级命令
	l.addCommand(newChannelCMD(ctx))    // 频道命令
	l.addCommand(newSubscriberCMD(ctx)) // 订阅者命令
	l.addCommand(newMockCMD(ctx))       // mock命令
	l.addCommand(newUserCMD(ctx))       // 用户命令
	l.addCommand(newDenylistCMD(ctx))   // 黑名单命令
	l.addCommand(newAllowlistCMD(ctx))  // 白名单命令

	if err := l.rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
