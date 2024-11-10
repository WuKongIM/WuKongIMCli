package cmd

import (
	"log"
	"strconv"

	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
)

type subscriberVar struct {
	chType    int // 频道类型
	list      []string
	chPrefix  string // 频道前缀
	chNum     int    // 频道数量
	subNum    int    // 订阅者数量
	subPrefix string // 订阅者前缀
}

type subscriberCMD struct {
	ctx           *WuKongIMContext
	api           *API
	subscriberVar *subscriberVar
}

func newSubscriberCMD(ctx *WuKongIMContext) *subscriberCMD {
	s := &subscriberCMD{
		ctx:           ctx,
		api:           NewAPI(),
		subscriberVar: &subscriberVar{},
	}
	return s
}

func (s *subscriberCMD) CMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscriber",
		Short: "subscriber",
	}

	add := &cobra.Command{
		Use:   "add",
		Short: "add subscriber",
		RunE:  s.runAdd,
	}

	remove := &cobra.Command{
		Use:   "remove",
		Short: "remove subscriber",
		RunE:  s.runRemove,
	}

	cmd.AddCommand(add)
	cmd.AddCommand(remove)

	s.initSubscriberVar(add)
	s.initSubscriberVar(remove)

	return cmd
}

func (s *subscriberCMD) initSubscriberVar(cmd *cobra.Command) {
	cmd.Flags().StringVar(&s.subscriberVar.chPrefix, "chPrefix", "ch", "频道前缀")
	cmd.Flags().IntVar(&s.subscriberVar.chType, "chType", 2, "频道类型")
	cmd.Flags().StringSliceVar(&s.subscriberVar.list, "list", []string{}, "频道订阅者列表")
	cmd.Flags().StringVar(&s.subscriberVar.subPrefix, "subPrefix", "sub", "订阅者前缀")
	cmd.Flags().IntVar(&s.subscriberVar.chNum, "chNum", 1, "频道数量")
	cmd.Flags().IntVar(&s.subscriberVar.subNum, "subNum", 0, "订阅者数量")
}

func (s *subscriberCMD) runAdd(cmd *cobra.Command, args []string) error {

	s.api.SetBaseURL(s.ctx.opts.ServerAddr)

	uiprogress.Start()
	defer uiprogress.Stop()

	progress := uiprogress.AddBar(s.subscriberVar.chNum).AppendCompleted().PrependElapsed()

	progress.Width = progressWidth()

	log.Printf("Starting subscriber add")

	state := "Setup     "

	progress.PrependFunc(func(b *uiprogress.Bar) string {
		return state
	})

	channels := make([]Channel, s.subscriberVar.chNum)
	for i := 0; i < s.subscriberVar.chNum; i++ {
		ch := Channel{
			ChannelId:   s.subscriberVar.chPrefix + strconv.Itoa(i),
			ChannelType: uint8(s.subscriberVar.chType),
		}
		channels[i] = ch
	}

	subscribers := make([]string, 0)
	if len(s.subscriberVar.list) > 0 {
		subscribers = append(subscribers, s.subscriberVar.list...)
	} else if s.subscriberVar.subNum > 0 {
		for i := 0; i < s.subscriberVar.subNum; i++ {
			subscribers = append(subscribers, s.subscriberVar.subPrefix+strconv.Itoa(i))
		}
	}

	for _, ch := range channels {
		err := s.api.SubscriberAdd(&SubscriberAddReq{
			ChannelId:   ch.ChannelId,
			ChannelType: ch.ChannelType,
			Subscribers: subscribers,
		})
		if err != nil {
			return err
		}
		progress.Incr()
	}

	state = "Finished  "

	return nil
}

func (s *subscriberCMD) runRemove(cmd *cobra.Command, args []string) error {
	s.api.SetBaseURL(s.ctx.opts.ServerAddr)

	uiprogress.Start()
	defer uiprogress.Stop()

	progress := uiprogress.AddBar(s.subscriberVar.chNum).AppendCompleted().PrependElapsed()

	progress.Width = progressWidth()

	log.Printf("Starting subscriber remove")

	state := "Setup     "

	progress.PrependFunc(func(b *uiprogress.Bar) string {
		return state
	})

	channels := make([]Channel, s.subscriberVar.chNum)
	for i := 0; i < s.subscriberVar.chNum; i++ {
		ch := Channel{
			ChannelId:   s.subscriberVar.chPrefix + strconv.Itoa(i),
			ChannelType: uint8(s.subscriberVar.chType),
		}
		channels[i] = ch
	}

	subscribers := make([]string, 0)
	if len(s.subscriberVar.list) > 0 {
		subscribers = append(subscribers, s.subscriberVar.list...)
	} else if s.subscriberVar.subNum > 0 {
		for i := 0; i < s.subscriberVar.subNum; i++ {
			subscribers = append(subscribers, s.subscriberVar.subPrefix+strconv.Itoa(i))
		}
	}

	for _, ch := range channels {
		err := s.api.SubscriberRemove(&SubscriberReq{
			ChannelId:   ch.ChannelId,
			ChannelType: ch.ChannelType,
			Subscribers: subscribers,
		})
		if err != nil {
			return err
		}
		progress.Incr()
	}

	state = "Finished  "

	return nil
}
