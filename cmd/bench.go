package cmd

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/WuKongIM/WuKongIM/pkg/client"
	"github.com/WuKongIM/WuKongIMCli/bench"
	wkproto "github.com/WuKongIM/WuKongIMGoProto"
	"github.com/dustin/go-humanize"
	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
)

type benchCMD struct {
	channels    []string
	channelType uint8
	channelNum  int // 频道数量
	pub         int
	sub         int
	msgs        int
	msgSize     int
	ctx         *WuKongIMContext
	noProgress  bool
	channel     *client.Channel
	channelList []*client.Channel
	pubSleep    time.Duration
	p2p         bool   // 是否是点对点聊天
	fromUID     string // 如果是p2p模式 则对应的发送者
	toUID       string // 如果是p2p模式 则对应的接受者
	api         *API
}

func newBenchCMD(ctx *WuKongIMContext) *benchCMD {
	b := &benchCMD{
		ctx: ctx,
		api: NewAPI(),
	}
	return b
}

func (b *benchCMD) CMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bench",
		Short: "stress testing",
		RunE:  b.run,
	}
	b.initVar(cmd)
	return cmd
}

func (b *benchCMD) initVar(cmd *cobra.Command) {
	cmd.Flags().IntVar(&b.pub, "pub", 0, "Number of concurrent senders(发送者数量)")
	cmd.Flags().IntVar(&b.sub, "sub", 0, "Number of concurrent receiver（接受者数量")
	cmd.Flags().IntVar(&b.msgs, "msgs", 100000, "Number of messages to publish（消息数量）")
	cmd.Flags().IntVar(&b.msgSize, "size", 128, "Size of the test messages,unit byte（测试消息大小,单位byte）")
	cmd.Flags().BoolVar(&b.noProgress, "no-progress", false, "Disable progress bar while publishing（不显示进度条）")
	cmd.Flags().DurationVar(&b.pubSleep, "pubsleep", 0, "Sleep for the specified interval after publishing each message（每条消息发送间隔）")
	cmd.Flags().StringArrayVar(&b.channels, "channels", []string{}, "channel list（接受消息的频道集合）")
	cmd.Flags().Uint8Var(&b.channelType, "channelType", 6, "channel type（频道类型）")
	cmd.Flags().IntVar(&b.channelNum, "channelNum", 1, "channel number（频道数量）")

}

func (b *benchCMD) run(cmd *cobra.Command, args []string) error {
	b.api.SetBaseURL(b.ctx.opts.ServerAddr)

	if b.channelType == 0 {
		b.channelType = 6
	}

	// ========== 初始化频道 ==========
	if len(b.channels) > 0 {
		for _, channel := range b.channels {
			b.channelList = append(b.channelList, client.NewChannel(channel, b.channelType))
		}
	}
	if b.channelNum > 0 {
		for i := 0; i < b.channelNum; i++ {
			b.channelList = append(b.channelList, client.NewChannel(fmt.Sprintf("channel-%04d", i), b.channelType))
		}
	}

	// ========== 生成用户uid ==========
	userPrefix := strconv.FormatInt(time.Now().UnixMilli(), 16) // 客户端前缀
	publishers := []string{}                                    // 发布者
	subscribers := []string{}                                   // 订阅者
	for i := 0; i < b.pub; i++ {
		uid := fmt.Sprintf("%s-%d", userPrefix, i)
		publishers = append(publishers, uid)
	}

	for i := 0; i < b.sub; i++ {
		uid := fmt.Sprintf("%s-%d", userPrefix, i)
		subscribers = append(subscribers, uid)
	}

	// ========== 获取用户的长连接地址 ==========
	userTcpAddrMap, err := b.api.Route(append(publishers, subscribers...))
	if err != nil {
		panic(err)
	}

	// ========== 创建客户端 ==========
	pubClients := make([]*client.Client, 0)
	subClients := make([]*client.Client, 0)
	for _, uid := range publishers {
		tcpAddr := userTcpAddrMap[uid]
		cli := client.New(tcpAddr, client.WithUID(uid))
		pubClients = append(pubClients, cli)
	}
	for _, uid := range subscribers {
		tcpAddr := userTcpAddrMap[uid]
		cli := client.New(tcpAddr, client.WithUID(uid))
		subClients = append(subClients, cli)
	}

	// ========== 连接客户端 ==========
	startwg := &sync.WaitGroup{}
	donewg := &sync.WaitGroup{}
	trigger := make(chan struct{})

	clientNum := len(pubClients) + len(subClients)

	uiprogress.Start()

	progress := uiprogress.AddBar(clientNum).AppendCompleted().PrependElapsed()
	state := "Connecting"
	progress.PrependFunc(func(b *uiprogress.Bar) string {
		return state
	})
	progress.Width = progressWidth()

	log.Printf("Starting WuKongIM  pub/sub benchmark [msgSize=%s]", humanize.IBytes(uint64(b.msgSize)))
	log.Printf("Connecting..., %s clients", humanize.Comma(int64(clientNum)))
	for _, cli := range pubClients {
		err = cli.Connect()
		if err != nil {
			log.Fatalf("Could not connect to the server: %v", err)
		}
		progress.Incr()
	}
	for _, cli := range subClients {
		err = cli.Connect()
		if err != nil {
			log.Fatalf("Could not connect to the server: %v", err)
		}
		progress.Incr()
	}
	state = "Connected "

	// ========== 接受消息 ==========

	bm := bench.NewBenchmark("WuKongIM", b.sub, b.pub)
	for _, cli := range subClients {
		startwg.Add(1)
		donewg.Add(1)
		go b.runReceiver(bm, cli, startwg, donewg, b.msgs)
	}
	startwg.Wait()

	// ========== 发送消息 ==========
	for _, cli := range pubClients {
		startwg.Add(1)
		donewg.Add(1)
		go b.runSender(bm, cli, startwg, donewg, trigger, b.msgs)
	}

	startwg.Wait()
	close(trigger)
	donewg.Wait()
	bm.Close()

	uiprogress.Stop()

	fmt.Println()

	fmt.Println(bm.Report())
	return nil
}

func (b *benchCMD) runReceiver(bm *bench.Benchmark, cli *client.Client, startwg *sync.WaitGroup, donewg *sync.WaitGroup, numMsg int) {
	received := 0

	ch := make(chan time.Time, 2)

	var progress *uiprogress.Bar

	log.Printf("Starting receiver, expecting %s messages", humanize.Comma(int64(numMsg)))

	if !b.noProgress {
		progress = uiprogress.AddBar(numMsg).AppendCompleted().PrependElapsed()
		progress.Width = progressWidth()
	}
	state := "Setup     "

	if progress != nil {
		progress.PrependFunc(func(b *uiprogress.Bar) string {
			return state
		})
	}
	messageHandler := func(msg *wkproto.RecvPacket) error {
		received++
		if received == 1 {
			ch <- time.Now()
		}
		if received >= numMsg {
			ch <- time.Now()
		}
		if progress != nil {
			progress.Incr()
		}
		return nil
	}
	state = "Receiving "

	cli.SetOnRecv(messageHandler)

	startwg.Done()

	start := <-ch
	end := <-ch
	state = "Finished  "

	bm.AddSubSample(bench.NewSample(numMsg, b.msgSize, start, end, cli))

	donewg.Done()
}

func (b *benchCMD) runSender(bm *bench.Benchmark, cli *client.Client, startwg *sync.WaitGroup, donewg *sync.WaitGroup, trigger chan struct{}, numMsg int) {

	startwg.Done()
	var progress *uiprogress.Bar

	log.Printf("Starting pub, sending %s messages", humanize.Comma(int64(numMsg)))

	if !b.noProgress {
		progress = uiprogress.AddBar(numMsg).AppendCompleted().PrependElapsed()
		progress.Width = progressWidth()
	}
	var msg []byte
	if b.msgSize > 0 {
		msg = make([]byte, b.msgSize)
	}
	<-trigger

	start := time.Now()
	var finishWg = &sync.WaitGroup{}
	b.publisher(cli, progress, msg, numMsg, finishWg)
	err := cli.Flush()
	if err != nil {
		log.Fatalf("Could not flush the connection: %v", err)
	}
	finishWg.Wait()
	bm.AddPubSample(bench.NewSample(numMsg, b.msgSize, start, time.Now(), cli))

	donewg.Done()
}

func (b *benchCMD) publisher(cli *client.Client, progress *uiprogress.Bar, msg []byte, numMsg int, finishWg *sync.WaitGroup) {

	state := "Sending"
	var err error
	if progress != nil {
		progress.PrependFunc(func(b *uiprogress.Bar) string {
			return state
		})
	}

	cli.SetOnSendack(func(sendackPacket *wkproto.SendackPacket) {
		if progress != nil {
			progress.Incr()
		}
		finishWg.Done()
	})
	for i := 0; i < numMsg; i++ {

		finishWg.Add(1)
		err = cli.SendMessage(b.channelList[i%len(b.channelList)], msg, client.SendOptionWithNoEncrypt(false))
		if err != nil {
			log.Fatalf("SendMessage error: %v", err)
		}
		time.Sleep(b.pubSleep)
	}
	state = "Finished  "

}
