package cmd

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/WuKongIM/WuKongIM/pkg/client"
	"github.com/WuKongIM/WuKongIM/pkg/wkproto"
	"github.com/WuKongIM/WuKongIMCli/bench"
	"github.com/dustin/go-humanize"
	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
)

type benchCMD struct {
	channelStr string
	sender     int
	receiver   int
	msgs       int
	msgSize    int
	ctx        *WuKongIMContext
	noProgress bool
	channel    *client.Channel
	pubSleep   time.Duration
	p2p        bool   // 是否是点对点聊天
	fromUID    string // 如果是p2p模式 则对应的发送者
	toUID      string // 如果是p2p模式 则对应的接受者
}

func newBenchCMD(ctx *WuKongIMContext) *benchCMD {
	b := &benchCMD{
		ctx: ctx,
	}
	return b
}

func (b *benchCMD) CMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bench",
		Short: "压力测试",
		RunE:  b.run,
	}
	b.initVar(cmd)
	return cmd
}

func (b *benchCMD) initVar(cmd *cobra.Command) {
	cmd.Flags().IntVar(&b.sender, "s", 0, "Number of concurrent senders(发送者数量)")
	cmd.Flags().IntVar(&b.receiver, "r", 0, "Number of concurrent receiver（接受者数量")
	cmd.Flags().IntVar(&b.msgs, "msgs", 100000, "Number of messages to publish（消息数量）")
	cmd.Flags().IntVar(&b.msgSize, "size", 128, "Size of the test messages,unit byte（测试消息大小,单位byte）")
	cmd.Flags().BoolVar(&b.noProgress, "no-progress", false, "Disable progress bar while publishing（不显示进度条）")
	cmd.Flags().DurationVar(&b.pubSleep, "pubsleep", 0, "Sleep for the specified interval after publishing each message（每条消息发送间隔）")

}

func (b *benchCMD) run(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		cmd.Help()
		return errors.New("channel is empty！")
	}
	b.channelStr = args[0]
	if strings.Contains(b.channelStr, "@") {
		channels := strings.Split(b.channelStr, "@")
		if len(channels) != 2 {
			cmd.Help()
			return nil
		}
		b.p2p = true
		b.fromUID = channels[0]
		b.toUID = channels[1]
		b.sender = 1
		b.receiver = 1
	}
	if b.sender == 0 {
		cmd.Help()
		return nil
	}
	if b.p2p {
		b.channel = client.NewChannel(b.toUID, 1)
	} else {
		b.channel = client.NewChannel(b.channelStr, 6)
	}

	var offset = func(putter int, counts []int) int {
		var position = 0

		for i := 0; i < putter; i++ {
			position = position + counts[i]
		}
		return position
	}

	log.Printf("Starting WuKongIM  send/recv benchmark [msgSize=%s]", humanize.IBytes(uint64(b.msgSize)))

	startwg := &sync.WaitGroup{}
	donewg := &sync.WaitGroup{}
	trigger := make(chan struct{})
	benchId := strconv.FormatInt(time.Now().UnixMilli(), 16) // 客户端前缀
	bm := bench.NewBenchmark("WuKongIM", b.receiver, b.sender)
	senderCounts := bench.MsgsPerClient(b.msgs, b.sender) // 每个客户端发送消息数量

	subCounts := bench.MsgsPerClient(b.msgs, b.receiver)

	for i := 0; i < b.receiver; i++ {
		uid := fmt.Sprintf("receiver-%s-%d-%d", benchId, i, i+offset(i, subCounts))
		if b.p2p {
			uid = b.toUID
		}
		cli := client.New(b.ctx.opts.ServerAddr, client.WithUID(uid))
		defer cli.Close()
		err := cli.Connect()
		if err != nil {
			return fmt.Errorf("WuKongIM connection %d failed: %s", i, err)
		}

		startwg.Add(1)
		donewg.Add(1)

		go b.runReceiver(bm, cli, startwg, donewg, b.msgs)
	}

	startwg.Wait()

	for i := 0; i < b.sender; i++ {
		uid := fmt.Sprintf("%s-%d-%d", benchId, i, i+offset(i, senderCounts))
		if b.p2p {
			uid = b.fromUID
		}
		cli := client.New(b.ctx.opts.ServerAddr, client.WithUID(uid))
		defer cli.Close()

		err := cli.Connect()
		if err != nil {
			return fmt.Errorf("WuKongIM connection %d failed: %s", i, err)
		}

		startwg.Add(1)
		donewg.Add(1)

		go b.runSender(bm, cli, startwg, donewg, trigger, senderCounts[i])
	}
	if !b.noProgress {
		uiprogress.Start()
	}

	startwg.Wait()
	close(trigger)
	donewg.Wait()

	bm.Close()

	if !b.noProgress {
		uiprogress.Stop()
	}
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

	log.Printf("Starting sender, sending %s messages", humanize.Comma(int64(numMsg)))

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
		err = cli.SendMessage(b.channel, msg, client.SendOptionWithNoEncrypt(false))
		if err != nil {
			log.Fatalf("SendMessage error: %v", err)
		}
		time.Sleep(b.pubSleep)
	}
	state = "Finished  "

}
