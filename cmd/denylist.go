package cmd

import (
	"log"
	"strconv"

	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
)

type denylistVar struct {
	chType    int // 频道类型
	list      []string
	chPrefix  string // 频道前缀
	chNum     int    // 频道数量
	subNum    int    // 订阅者数量
	subPrefix string // 订阅者前缀
}

type denylistCMD struct {
	ctx         *WuKongIMContext
	api         *API
	denylistVar *denylistVar
}

func newDenylistCMD(ctx *WuKongIMContext) *denylistCMD {
	s := &denylistCMD{
		ctx:         ctx,
		api:         NewAPI(),
		denylistVar: &denylistVar{},
	}
	return s
}

func (s *denylistCMD) CMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "denylist",
		Short: "denylist",
	}

	add := &cobra.Command{
		Use:   "add",
		Short: "add denylist",
		RunE:  s.runAdd,
	}

	remove := &cobra.Command{
		Use:   "remove",
		Short: "remove denylist",
		RunE:  s.runRemove,
	}

	cmd.AddCommand(add)
	cmd.AddCommand(remove)

	s.initDenylistVar(add)
	s.initDenylistVar(remove)

	return cmd
}

func (s *denylistCMD) initDenylistVar(cmd *cobra.Command) {
	cmd.Flags().StringVar(&s.denylistVar.chPrefix, "chPrefix", "ch", "频道前缀")
	cmd.Flags().IntVar(&s.denylistVar.chType, "chType", 2, "频道类型")
	cmd.Flags().StringSliceVar(&s.denylistVar.list, "list", []string{}, "频道订阅者列表")
	cmd.Flags().StringVar(&s.denylistVar.subPrefix, "subPrefix", "sub", "订阅者前缀")
	cmd.Flags().IntVar(&s.denylistVar.chNum, "chNum", 1, "频道数量")
	cmd.Flags().IntVar(&s.denylistVar.subNum, "subNum", 0, "订阅者数量")
}

func (s *denylistCMD) runAdd(cmd *cobra.Command, args []string) error {

	s.api.SetBaseURL(s.ctx.opts.ServerAddr)

	uiprogress.Start()
	defer uiprogress.Stop()

	progress := uiprogress.AddBar(s.denylistVar.chNum).AppendCompleted().PrependElapsed()

	progress.Width = progressWidth()

	log.Printf("Starting denylist add")

	state := "Setup     "

	progress.PrependFunc(func(b *uiprogress.Bar) string {
		return state
	})

	channels := make([]Channel, s.denylistVar.chNum)
	for i := 0; i < s.denylistVar.chNum; i++ {
		ch := Channel{
			ChannelId:   s.denylistVar.chPrefix + strconv.Itoa(i),
			ChannelType: uint8(s.denylistVar.chType),
		}
		channels[i] = ch
	}

	subscribers := make([]string, 0)
	if len(s.denylistVar.list) > 0 {
		subscribers = append(subscribers, s.denylistVar.list...)
	} else if s.denylistVar.subNum > 0 {
		for i := 0; i < s.denylistVar.subNum; i++ {
			subscribers = append(subscribers, s.denylistVar.subPrefix+strconv.Itoa(i))
		}
	}

	for _, ch := range channels {
		err := s.api.DenylistAdd(&ChannelUidsReq{
			ChannelId:   ch.ChannelId,
			ChannelType: ch.ChannelType,
			Uids:        subscribers,
		})
		if err != nil {
			return err
		}
		progress.Incr()
	}

	state = "Finished  "

	return nil
}

func (s *denylistCMD) runRemove(cmd *cobra.Command, args []string) error {
	s.api.SetBaseURL(s.ctx.opts.ServerAddr)

	uiprogress.Start()
	defer uiprogress.Stop()

	progress := uiprogress.AddBar(s.denylistVar.chNum).AppendCompleted().PrependElapsed()

	progress.Width = progressWidth()

	log.Printf("Starting denylist remove")

	state := "Setup     "

	progress.PrependFunc(func(b *uiprogress.Bar) string {
		return state
	})

	channels := make([]Channel, s.denylistVar.chNum)
	for i := 0; i < s.denylistVar.chNum; i++ {
		ch := Channel{
			ChannelId:   s.denylistVar.chPrefix + strconv.Itoa(i),
			ChannelType: uint8(s.denylistVar.chType),
		}
		channels[i] = ch
	}

	subscribers := make([]string, 0)
	if len(s.denylistVar.list) > 0 {
		subscribers = append(subscribers, s.denylistVar.list...)
	} else if s.denylistVar.subNum > 0 {
		for i := 0; i < s.denylistVar.subNum; i++ {
			subscribers = append(subscribers, s.denylistVar.subPrefix+strconv.Itoa(i))
		}
	}

	for _, ch := range channels {
		err := s.api.DenylistRemove(&ChannelUidsReq{
			ChannelId:   ch.ChannelId,
			ChannelType: ch.ChannelType,
			Uids:        subscribers,
		})
		if err != nil {
			return err
		}
		progress.Incr()
	}

	state = "Finished  "

	return nil
}
