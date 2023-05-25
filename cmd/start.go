package cmd

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"
)

type startCMD struct {
	ctx         *WuKongIMContext
	installDir  string
	installName string
	sysos       string
	sysarch     string
	downloadUrl string
	pidfile     string

	version string
}

func newStartCMD(ctx *WuKongIMContext) *startCMD {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return &startCMD{
		ctx:         ctx,
		installDir:  path.Join(homeDir, "wukongim"),
		installName: "wukongim",
		pidfile:     "wukongim.lock",
		downloadUrl: "https://github.com/WuKongIM/WuKongIM/releases/download/${version}/wukongim-${sysos}-${sysarch}",
	}
}

func (s *startCMD) CMD() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start a Wukong IM service.",
		RunE:  s.run,
	}
	startCmd.Flags().StringVar(&s.version, "version", "v1.0.1", "Version number of Wukong IM")

	stopCMD := &cobra.Command{
		Use:   "stop",
		Short: "Stop a Wukong IM service.",
		Run: func(cmd *cobra.Command, args []string) {
			err := s.stop()
			if err != nil {
				fmt.Println("WukongIM stop failed.", err)
				return
			}
			println("WukongIM has stopped.")

		},
	}

	s.ctx.w.rootCmd.AddCommand(stopCMD)

	return startCmd
}

func (s *startCMD) run(cmd *cobra.Command, args []string) error {

	err := os.MkdirAll(s.installDir, 0755)
	if err != nil {
		return err
	}

	sysos := runtime.GOOS

	sysarch := runtime.GOARCH

	s.sysos = strings.ToLower(sysos)
	s.sysarch = strings.ToLower(sysarch)
	installPath := path.Join(s.installDir, s.installName)
	fmt.Println("Install Dir is " + s.installDir)
	if !s.execIsExist() {
		tmpPath, err := s.download()
		if err != nil {
			return err
		}
		err = os.Rename(tmpPath, installPath)
		if err != nil {
			return err
		}
	}
	if err := s.start(); err != nil {
		return err
	}

	fmt.Println("WukongIM started successfully.")

	return nil
}

func (s *startCMD) start() error {
	installPath := path.Join(s.installDir, s.installName)
	err := os.Chmod(installPath, 0755)
	if err != nil {
		return err
	}

	cm := exec.Command(installPath, flag.Args()...)
	if err := cm.Start(); err != nil {
		return err
	}
	return ioutil.WriteFile(path.Join(s.installDir, s.pidfile), []byte(strconv.Itoa(cm.Process.Pid)), 0644)

}

func (s *startCMD) stop() error {
	strb, _ := ioutil.ReadFile(path.Join(s.installDir, s.pidfile))
	command := exec.Command("kill", string(strb))
	err := command.Start()
	return err
}

func (s *startCMD) execIsExist() bool {
	installPath := path.Join(s.installDir, s.installName)
	_, err := os.Stat(installPath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false

}

func (s *startCMD) download() (string, error) {

	downloadURL := strings.ReplaceAll(s.downloadUrl, "${version}", s.version)
	downloadURL = strings.ReplaceAll(downloadURL, "${sysos}", s.sysos)
	downloadURL = strings.ReplaceAll(downloadURL, "${sysarch}", s.sysarch)

	fmt.Println("Start download wukongim from " + downloadURL + " ...")

	destPath := path.Join(os.TempDir(), "wukongim_tmp")

	client := http.DefaultClient
	client.Timeout = 60 * 10 * time.Second
	reps, err := client.Get(downloadURL)
	if err != nil {
		return "", err
	}
	if reps.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http status[%d] is error", reps.StatusCode)
	}
	//保存文件
	file, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer file.Close() //关闭文件

	//获取下载文件的大小
	length := reps.Header.Get("Content-Length")
	size, _ := strconv.ParseInt(length, 10, 64)
	body := reps.Body //获取文件内容
	bar := pb.Full.Start64(size)
	bar.SetWidth(120)                         //设置进度条宽度
	bar.SetRefreshRate(10 * time.Millisecond) //设置刷新速率
	defer bar.Finish()
	// create proxy reader
	barReader := bar.NewProxyReader(body)
	//写入文件
	writer := io.Writer(file)
	io.Copy(writer, barReader)

	return destPath, nil

}
