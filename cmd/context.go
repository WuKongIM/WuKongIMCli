package cmd

import (
	"errors"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type contextCMD struct {
	cmd         *cobra.Command
	description string
	server      string
	token       string
	ctx         *WuKongIMContext
}

func newContextCMD(ctx *WuKongIMContext) *contextCMD {
	c := &contextCMD{
		ctx: ctx,
	}
	c.cmd = &cobra.Command{
		Use:   "context",
		Short: "Manage WuKongIM configuration contexts（管理WuKongIM的配置信息）",
	}
	return c
}

func (c *contextCMD) CMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Manage WuKongIM configuration contexts（管理WuKongIM的配置信息）",
	}
	c.initSubCMD(cmd)
	return cmd
}

func (c *contextCMD) initSubCMD(cmd *cobra.Command) {
	addCMD := &cobra.Command{
		Use:   "add",
		Short: "Update or create a context（更新或创建一个狸猫IM的上下文）",
		RunE:  c.add,
	}
	addCMD.Flags().StringVar(&c.description, "description", "", "Context description （上下文的描述）")
	addCMD.Flags().StringVarP(&c.server, "server", "s", "tcp://127.0.0.1:7677", "Server address（WuKongIM服务器地址）")
	addCMD.Flags().StringVar(&c.token, "token", "", "Token for connect WuKongIM （连接WuKongIM的token）")
	cmd.AddCommand(addCMD)
}

func (c *contextCMD) add(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		cmd.Help()
		return nil
	}
	name := args[0]
	if !validName(name) {
		return errors.New("无效的名字")
	}
	c.ctx.opts.Description = c.description
	c.ctx.opts.ServerAddr = c.server
	c.ctx.opts.Token = c.token
	return c.ctx.opts.Save(name)
}

func validName(name string) bool {
	return name != "" && !strings.Contains(name, "..") && !strings.Contains(name, string(os.PathSeparator))
}
