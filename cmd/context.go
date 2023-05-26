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
		Short: "Manage WuKongIM configuration contexts",
	}
	return c
}

func (c *contextCMD) CMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Manage WuKongIM configuration contexts",
	}
	c.initSubCMD(cmd)
	return cmd
}

func (c *contextCMD) initSubCMD(cmd *cobra.Command) {
	addCMD := &cobra.Command{
		Use:   "add",
		Short: "Update or create a context",
		RunE:  c.add,
	}
	addCMD.Flags().StringVar(&c.description, "description", c.ctx.opts.Description, "Context description")
	addCMD.Flags().StringVarP(&c.server, "server", "s", c.ctx.opts.ServerAddr, "Http  api server address")
	addCMD.Flags().StringVar(&c.token, "token", "", "Token for connect WuKongIM")
	cmd.AddCommand(addCMD)
}

func (c *contextCMD) add(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		cmd.Help()
		return nil
	}
	name := args[0]
	if !validName(name) {
		return errors.New("invalid name")
	}
	c.ctx.opts.Description = c.description
	c.ctx.opts.ServerAddr = c.server
	c.ctx.opts.Token = c.token
	return c.ctx.opts.Save(name)
}

func validName(name string) bool {
	return name != "" && !strings.Contains(name, "..") && !strings.Contains(name, string(os.PathSeparator))
}
