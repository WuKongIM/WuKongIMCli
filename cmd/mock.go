package cmd

import (
	"errors"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/WuKongIM/WuKongIM/pkg/client"
	wkproto "github.com/WuKongIM/WuKongIMGoProto"
	"github.com/gosuri/uiprogress"
	"github.com/panjf2000/ants/v2"
	"github.com/spf13/cobra"
)

type mockVar struct {
	num      int           // 用户或群的数量
	prefix   string        // 用户前缀
	chPrefix string        // 频道前缀
	chType   int           // 频道类型
	chNum    int           // 频道数量
	interval time.Duration // 发送消息间隔
	duration time.Duration // 程序持续时间

}

type mockCMD struct {
	ctx           *WuKongIMContext
	api           *API
	mockVar       *mockVar
	userClientMap map[string]*testClient // 在线用户client
	uids          []string               // 在线用户uids

	channels []Channel
}

func newMockCMD(ctx *WuKongIMContext) *mockCMD {
	s := &mockCMD{
		ctx:           ctx,
		api:           NewAPI(),
		mockVar:       &mockVar{},
		userClientMap: make(map[string]*testClient),
	}
	return s
}

func (m *mockCMD) CMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mock",
		Short: "mock",
	}

	// 在线命令
	online := &cobra.Command{
		Use:   "online",
		Short: "online user",
		RunE:  m.runOnline,
	}

	// 聊天命令
	chat := &cobra.Command{
		Use:   "chat",
		Short: "chat",
		RunE:  m.runChat,
	}

	cmd.AddCommand(online)
	cmd.AddCommand(chat)

	m.initMockVar(online)
	m.initMockVar(chat)

	return cmd
}

func (m *mockCMD) initMockVar(cmd *cobra.Command) {
	cmd.Flags().StringVar(&m.mockVar.prefix, "prefix", "usr", "user prefix")
	cmd.Flags().StringVar(&m.mockVar.chPrefix, "chPrefix", "", "channel prefix")
	cmd.Flags().IntVar(&m.mockVar.num, "num", 0, "user number")
	cmd.Flags().DurationVar(&m.mockVar.duration, "duration", 0, "duration")
	cmd.Flags().IntVar(&m.mockVar.chType, "chType", 2, "channel type")
	cmd.Flags().IntVar(&m.mockVar.chNum, "chNum", 1, "channel number")
	cmd.Flags().DurationVar(&m.mockVar.interval, "interval", 5, "interval")
}

func (m *mockCMD) runOnline(cmd *cobra.Command, args []string) error {

	// create progress bar
	uiprogress.Start()

	var progress *uiprogress.Bar
	progressNum := m.mockVar.num
	if m.mockVar.num <= 0 {
		progressNum = 1
	}
	progress = uiprogress.AddBar(progressNum).AppendCompleted().PrependElapsed()

	progress.Width = progressWidth()

	log.Printf("Starting user online")

	state := "Setup     "

	progress.PrependFunc(func(b *uiprogress.Bar) string {
		return state
	})

	// online user
	err := m.onlineUser(m.mockVar.num, func(cli *testClient) {
		progress.Incr()
		if progress.Current() >= progress.Total {
			state = "Done  "
			uiprogress.Stop()
		}
	})
	if err != nil {
		return err
	}

	tk := time.NewTicker(10 * time.Second)

	exit := time.After(m.mockVar.duration)
	if m.mockVar.duration == 0 {
		exit = time.After(time.Hour * 24 * 365 * 1) //  1 years, no exit
	}
	for {
		select {
		case <-tk.C:
			m.printOnline()
		case <-exit:
			return nil
		}
	}
}

func (m *mockCMD) runChat(cmd *cobra.Command, args []string) error {

	if m.mockVar.num <= 1 {
		return errors.New("user number must be greater than 1")
	}

	// create progress bar
	uiprogress.Start()

	var progress *uiprogress.Bar
	progressNum := m.mockVar.num
	if m.mockVar.num <= 0 {
		progressNum = 1
	}
	progress = uiprogress.AddBar(progressNum).AppendCompleted().PrependElapsed()

	progress.Width = progressWidth()

	log.Printf("Starting user online")

	state := "Setup     "

	progress.PrependFunc(func(b *uiprogress.Bar) string {
		return state
	})

	// online user
	err := m.onlineUser(m.mockVar.num, func(cli *testClient) {
		progress.Incr()
		if progress.Current() >= progress.Total {
			state = "Done  "
			uiprogress.Stop()
		}

		cli.SetOnRecv(func(recv *wkproto.RecvPacket) error {
			return m.onRecvMessage(cli, recv)
		})
	})
	if err != nil {
		return err
	}

	// init channel if need
	if m.mockVar.chPrefix != "" {
		for i := 0; i < m.mockVar.chNum; i++ {
			ch := Channel{
				ChannelId:   m.mockVar.chPrefix + strconv.Itoa(i),
				ChannelType: uint8(m.mockVar.chType),
			}
			m.channels = append(m.channels, ch)
		}
	}

	m.runSender()

	return nil
}

func (m *mockCMD) runSender() {
	// random send message
	var (
		err       error
		toChannel *Channel
	)
	tk := time.NewTicker(time.Second)

	timeout := m.mockVar.duration
	if timeout == 0 {
		timeout = time.Hour * 24 * 365 * 1
	}
	for {
		select {
		case <-tk.C:
			for _, uid := range m.uids {
				cli := m.userClientMap[uid]
				if cli.GetPreTo() != nil {
					toChannel = cli.GetPreTo()
				} else if len(m.channels) > 0 {
					toChannel = &m.channels[rand.Intn(len(m.channels))]
				} else {
					toChannel = &Channel{
						ChannelId:   m.getRandomUser(uid),
						ChannelType: uint8(m.mockVar.chType),
					}
				}
				err = cli.SendMessageToIfNeed(toChannel)
				if err != nil {
					log.Printf("send message error: %s", err)
				}
			}
		case <-time.After(timeout):
			return
		}
	}
}

func (m *mockCMD) onRecvMessage(cli *testClient, recv *wkproto.RecvPacket) error {
	cli.RecvInc(recv)
	return nil
}

func (m *mockCMD) onlineUser(num int, callback func(cli *testClient)) error {
	m.api.SetBaseURL(m.ctx.opts.ServerAddr)
	// generate user id
	m.uids = make([]string, num)
	for i := 0; i < num; i++ {
		m.uids[i] = m.mockVar.prefix + strconv.Itoa(i)
	}
	m.userClientMap = make(map[string]*testClient)

	// get user tcp addr
	userTcpAddrMap, err := m.api.Route(m.uids)
	if err != nil {
		return err
	}
	// create user client
	pool, err := ants.NewPoolWithFunc(20, func(cliObj interface{}) {
		cli := cliObj.(*testClient)
		callback(cli)
		err = cli.Connect()
		if err != nil {
			log.Printf("connect error: %s", err)
			return
		}

	})
	if err != nil {
		return err
	}
	defer pool.Release()

	for _, uid := range m.uids {
		tcpAddr := userTcpAddrMap[uid]
		cli := client.New(tcpAddr, client.WithUID(uid), client.WithAutoReconn(true))
		testCli := newTestClient(cli, m.mockVar.interval)
		m.userClientMap[uid] = testCli

		err = pool.Invoke(testCli)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *mockCMD) getRandomUser(uid string) string {
	i := rand.Intn(len(m.uids))
	if m.uids[i] == uid {
		return m.getRandomUser(uid)
	}
	return m.uids[i]
}

func (m *mockCMD) printOnline() {

	onlineCount := 0
	for _, uid := range m.uids {
		cli := m.userClientMap[uid]
		if cli.IsConnected() {
			onlineCount++
		}
	}

	log.Printf("online user count: %d \n", onlineCount)
}
