package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"sync"
	"time"

	"github.com/WuKongIM/WuKongIM/pkg/client"
	wkproto "github.com/WuKongIM/WuKongIMGoProto"
	"github.com/spf13/cobra"
	"go.uber.org/atomic"
	terminal "golang.org/x/term"
)

type CMD interface {
	CMD() *cobra.Command
}

type Options struct {
	ServerAddr  string
	Description string
	Token       string
}

func NewOptions() *Options {
	opts := &Options{
		ServerAddr:  "http://127.0.0.1:5001",
		Description: "",
	}
	err := os.MkdirAll(opts.ContextDir(), 0700)
	if err != nil {
		panic(err)
	}
	return opts
}

func (o *Options) ContextPath(name string) (string, error) {
	if !validName(name) {
		return "", fmt.Errorf("invalid context name %q", name)
	}
	return filepath.Join(o.ContextDir(), name+".json"), nil
}

func (o *Options) ContextDir() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	if u.HomeDir == "" {
		return ""
	}
	return filepath.Join(u.HomeDir, "wukongim", ".config", "context")

}

func (o *Options) Load() error {
	data, err := ioutil.ReadFile(o.metaFile())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(data) == 0 {
		return nil
	}
	name := string(data)
	filen, err := o.ContextPath(name)
	if err != nil {
		return err
	}

	optionData, err := ioutil.ReadFile(filen)
	if err != nil {
		return err
	}
	var optionMap map[string]interface{}
	err = json.Unmarshal(optionData, &optionMap)
	if err != nil {
		return err
	}
	if optionMap == nil {
		return nil
	}
	if optionMap["url"] != nil {
		o.ServerAddr = optionMap["url"].(string)
	}
	if optionMap["description"] != nil {
		o.Description = optionMap["description"].(string)
	}
	if optionMap["token"] != nil {
		o.Token = optionMap["token"].(string)
	}
	return nil
}

func (o *Options) Save(name string) error {
	p, err := o.ContextPath(name)
	if err != nil {
		return err
	}
	j, err := json.MarshalIndent(map[string]interface{}{
		"url":         o.ServerAddr,
		"description": o.Description,
		"token":       o.Token,
	}, "", "  ")
	if err != nil {
		return err
	}

	pf, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	_, err = pf.Write(j)
	if err != nil {
		return err
	}

	mf, err := os.OpenFile(o.metaFile(), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	_, err = mf.Write([]byte(name))
	return err
}

func (o *Options) metaFile() string {
	return filepath.Join(o.ContextDir(), "meta")
}

func progressWidth() int {
	w, _, err := terminal.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80
	}

	minWidth := 10

	if w-30 <= minWidth {
		return minWidth
	} else {
		return w - 30
	}
}

func move(oldPath, newPath string) error {
	srcFile, err := os.Open(oldPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(newPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	err = os.Remove(oldPath)
	if err != nil {
		return err
	}
	return nil
}

type testClient struct {
	cli       *client.Client
	sendSeq   atomic.Uint64
	interval  time.Duration
	lastSend  time.Time
	recvCount atomic.Uint64

	sendMap     map[string]uint64
	sendMapLock sync.Mutex
	sendLock    sync.Mutex

	recvMap     map[string]uint64
	recvMapLock sync.Mutex

	preTo *Channel // 上一次发送的目标
}

func newTestClient(cli *client.Client, interval time.Duration) *testClient {
	return &testClient{
		cli:      cli,
		interval: interval,
		recvMap:  make(map[string]uint64),
		sendMap:  make(map[string]uint64),
	}
}

func (t *testClient) IsConnected() bool {
	return t.cli.IsConnected()
}

func (t *testClient) Connect() error {
	return t.cli.Connect()
}

func (t *testClient) SendMessageToIfNeed(channel *Channel) error {
	t.sendLock.Lock()
	defer t.sendLock.Unlock()
	if time.Since(t.lastSend) < t.interval {
		return nil
	}
	t.lastSend = time.Now()
	t.sendMapLock.Lock()
	t.sendMap[channel.ChannelId]++
	t.sendMapLock.Unlock()
	t.preTo = channel
	return t.cli.SendMessage(client.NewChannel(channel.ChannelId, channel.ChannelType), []byte(fmt.Sprintf(`{"type":1,"content":"[%d]hello world!"}`, t.sendSeq.Inc())))
}

func (t *testClient) SetOnRecv(onRecvMessage func(recv *wkproto.RecvPacket) error) {
	t.cli.SetOnRecv(onRecvMessage)
}

func (t *testClient) RecvInc(recv *wkproto.RecvPacket) {
	t.recvCount.Inc()

	t.recvMapLock.Lock()
	defer t.recvMapLock.Unlock()
	t.recvMap[recv.FromUID]++
}

func (t *testClient) GetPreTo() *Channel {
	return t.preTo
}
