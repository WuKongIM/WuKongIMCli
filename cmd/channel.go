package cmd

import (
	"fmt"
	"log"

	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
)

type createVar struct {
	num    int    // 数量
	chType int    // 频道类型
	prefix string // 频道前缀
}

type channelCMD struct {
	ctx       *WuKongIMContext
	api       *API
	createVar *createVar
}

func newChannelCMD(ctx *WuKongIMContext) *channelCMD {
	c := &channelCMD{
		ctx:       ctx,
		api:       NewAPI(),
		createVar: &createVar{},
	}
	return c
}

func (c *channelCMD) CMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "channel",
		Short: "channel",
	}

	create := &cobra.Command{
		Use:   "create",
		Short: "create channel",
		RunE:  c.runCreate,
	}

	cmd.AddCommand(create)

	c.initCreateVar(create)

	return cmd
}

func (c *channelCMD) initCreateVar(cmd *cobra.Command) {
	cmd.Flags().StringVar(&c.createVar.prefix, "prefix", "", "频道前缀")
	cmd.Flags().IntVar(&c.createVar.num, "num", 0, "频道数量")
	cmd.Flags().IntVar(&c.createVar.chType, "chType", 2, "频道类型")
}

func (c *channelCMD) runCreate(cmd *cobra.Command, args []string) error {

	c.api.SetBaseURL(c.ctx.opts.ServerAddr)

	uiprogress.Start()
	defer uiprogress.Stop()
	var progress *uiprogress.Bar
	progressNum := c.createVar.num
	if c.createVar.num <= 0 {
		progressNum = 1
	}
	progress = uiprogress.AddBar(progressNum).AppendCompleted().PrependElapsed()

	progress.Width = progressWidth()

	log.Printf("Starting create channel")

	state := "Setup     "

	progress.PrependFunc(func(b *uiprogress.Bar) string {
		return state
	})

	if c.createVar.num <= 0 {
		err := c.api.CreateChannel(&ChannelCreateReq{
			ChannelInfoReq: ChannelInfoReq{
				ChannelId:   c.createVar.prefix,
				ChannelType: uint8(c.createVar.chType),
			},
		})
		if err != nil {
			return err
		}
		progress.Incr()
	} else {
		for i := 0; i < c.createVar.num; i++ {
			err := c.api.CreateChannel(&ChannelCreateReq{
				ChannelInfoReq: ChannelInfoReq{
					ChannelId:   fmt.Sprintf("%s%d", c.createVar.prefix, i),
					ChannelType: uint8(c.createVar.chType),
				},
			})
			if err != nil {
				return err
			}
			progress.Incr()
		}
	}
	state = "Finished  "

	return nil
}
