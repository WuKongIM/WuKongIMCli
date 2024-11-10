package cmd

import (
	"log"
	"strconv"

	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
)

type allowlistVar struct {
	chType    int // 频道类型
	list      []string
	chPrefix  string // 频道前缀
	chNum     int    // 频道数量
	subNum    int    // 订阅者数量
	subPrefix string // 订阅者前缀
}

type allowlistCMD struct {
	ctx          *WuKongIMContext
	api          *API
	allowlistVar *allowlistVar
}

func newAllowlistCMD(ctx *WuKongIMContext) *allowlistCMD {
	s := &allowlistCMD{
		ctx:          ctx,
		api:          NewAPI(),
		allowlistVar: &allowlistVar{},
	}
	return s
}

func (s *allowlistCMD) CMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "allowlist",
		Short: "allowlist",
	}

	add := &cobra.Command{
		Use:   "add",
		Short: "add allowlist",
		RunE:  s.runAdd,
	}

	remove := &cobra.Command{
		Use:   "remove",
		Short: "remove allowlist",
		RunE:  s.runRemove,
	}

	cmd.AddCommand(add)
	cmd.AddCommand(remove)

	s.initAllowlistVar(add)
	s.initAllowlistVar(remove)

	return cmd
}

func (s *allowlistCMD) initAllowlistVar(cmd *cobra.Command) {
	cmd.Flags().StringVar(&s.allowlistVar.chPrefix, "chPrefix", "ch", "频道前缀")
	cmd.Flags().IntVar(&s.allowlistVar.chType, "chType", 2, "频道类型")
	cmd.Flags().StringSliceVar(&s.allowlistVar.list, "list", []string{}, "频道订阅者列表")
	cmd.Flags().StringVar(&s.allowlistVar.subPrefix, "subPrefix", "sub", "订阅者前缀")
	cmd.Flags().IntVar(&s.allowlistVar.chNum, "chNum", 1, "频道数量")
	cmd.Flags().IntVar(&s.allowlistVar.subNum, "subNum", 0, "订阅者数量")
}

func (s *allowlistCMD) runAdd(cmd *cobra.Command, args []string) error {

	s.api.SetBaseURL(s.ctx.opts.ServerAddr)

	uiprogress.Start()
	defer uiprogress.Stop()

	progress := uiprogress.AddBar(s.allowlistVar.chNum).AppendCompleted().PrependElapsed()

	progress.Width = progressWidth()

	log.Printf("Starting denylist add")

	state := "Setup     "

	progress.PrependFunc(func(b *uiprogress.Bar) string {
		return state
	})

	channels := make([]Channel, s.allowlistVar.chNum)
	for i := 0; i < s.allowlistVar.chNum; i++ {
		ch := Channel{
			ChannelId:   s.allowlistVar.chPrefix + strconv.Itoa(i),
			ChannelType: uint8(s.allowlistVar.chType),
		}
		channels[i] = ch
	}

	subscribers := make([]string, 0)
	if len(s.allowlistVar.list) > 0 {
		subscribers = append(subscribers, s.allowlistVar.list...)
	} else if s.allowlistVar.subNum > 0 {
		for i := 0; i < s.allowlistVar.subNum; i++ {
			subscribers = append(subscribers, s.allowlistVar.subPrefix+strconv.Itoa(i))
		}
	}

	for _, ch := range channels {
		err := s.api.AllowlistAdd(&ChannelUidsReq{
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

func (s *allowlistCMD) runRemove(cmd *cobra.Command, args []string) error {
	s.api.SetBaseURL(s.ctx.opts.ServerAddr)

	uiprogress.Start()
	defer uiprogress.Stop()

	progress := uiprogress.AddBar(s.allowlistVar.chNum).AppendCompleted().PrependElapsed()

	progress.Width = progressWidth()

	log.Printf("Starting denylist remove")

	state := "Setup     "

	progress.PrependFunc(func(b *uiprogress.Bar) string {
		return state
	})

	channels := make([]Channel, s.allowlistVar.chNum)
	for i := 0; i < s.allowlistVar.chNum; i++ {
		ch := Channel{
			ChannelId:   s.allowlistVar.chPrefix + strconv.Itoa(i),
			ChannelType: uint8(s.allowlistVar.chType),
		}
		channels[i] = ch
	}

	subscribers := make([]string, 0)
	if len(s.allowlistVar.list) > 0 {
		subscribers = append(subscribers, s.allowlistVar.list...)
	} else if s.allowlistVar.subNum > 0 {
		for i := 0; i < s.allowlistVar.subNum; i++ {
			subscribers = append(subscribers, s.allowlistVar.subPrefix+strconv.Itoa(i))
		}
	}

	for _, ch := range channels {
		err := s.api.AllowlistRemove(&ChannelUidsReq{
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
