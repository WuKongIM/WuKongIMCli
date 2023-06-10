package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
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
	ctx                   *WuKongIMContext
	installDir            string
	installName           string
	sysos                 string
	sysarch               string
	downloadUrl           string
	chinaDownloadUrl      string
	configDowloadUrl      string // 配置下载地址
	chinaConfigDowloadUrl string
	configName            string
	pidfile               string

	version string
	china   bool
}

func newStartCMD(ctx *WuKongIMContext) *startCMD {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return &startCMD{
		ctx:                   ctx,
		installDir:            path.Join(homeDir, "wukongim"),
		installName:           "wukongim",
		pidfile:               "wukongim.lock",
		downloadUrl:           "https://github.com/WuKongIM/WuKongIM/releases/download/${version}/wukongim-${sysos}-${sysarch}",
		chinaDownloadUrl:      "https://githubim.com/wukongim/releases/${version}/wukongim-${sysos}-${sysarch}",
		configDowloadUrl:      "https://raw.githubusercontent.com/WuKongIM/WuKongIM/${version}/config/wk.yaml",
		chinaConfigDowloadUrl: "https://githubim.com/wukongim/releases/${version}/wk.yaml",
		configName:            "wk.yaml",
	}
}

func (s *startCMD) CMD() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start a WukongIM service.",
		RunE:  s.run,
	}
	startCmd.Flags().StringVar(&s.version, "version", "v1.0.5", "Version number of Wukong IM")
	startCmd.Flags().BoolVar(&s.china, "china", false, "china mirror")

	stopCMD := &cobra.Command{
		Use:   "stop",
		Short: "Stop a WukongIM service.",
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

	runCMD := &cobra.Command{
		Use:   "run",
		Short: "Run a Wukong IM service.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.runServer()
		},
	}
	s.ctx.w.rootCmd.AddCommand(runCMD)

	runCMD.Flags().StringVar(&s.version, "version", "v1.0.5", "Version number of WukongIM")
	runCMD.Flags().BoolVar(&s.china, "china", false, "china mirror")

	restartCMD := &cobra.Command{
		Use:   "restart",
		Short: "Restart a WukongIM service.",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := s.stop()
			if err != nil {
				return err
			}
			return s.start()
		},
	}
	s.ctx.w.rootCmd.AddCommand(restartCMD)

	return startCmd
}

func (s *startCMD) run(cmd *cobra.Command, args []string) error {

	if err := s.init(); err != nil {
		return err
	}

	if err := s.start(); err != nil {
		return err
	}

	fmt.Printf("Configuration file path is %s. \n", path.Join(s.installDir, s.configName))
	fmt.Println("WukongIM started successfully.")

	return nil
}

func (s *startCMD) init() error {
	err := os.MkdirAll(s.installDir, 0755)
	if err != nil {
		return err
	}

	sysos := runtime.GOOS

	sysarch := runtime.GOARCH

	s.sysos = strings.ToLower(sysos)
	s.sysarch = strings.ToLower(sysarch)
	fmt.Println("Installation directory is " + s.installDir)

	err = s.downloadIfNeed()
	if err != nil {
		return err
	}
	installPath := path.Join(s.installDir, s.installName)
	err = os.Chmod(installPath, 0755)
	if err != nil {
		return err
	}
	return nil
}

func (s *startCMD) start() error {
	installPath := path.Join(s.installDir, s.installName)
	cm := exec.Command(installPath, "--config", path.Join(s.installDir, s.configName))
	if err := cm.Start(); err != nil {
		return err
	}
	return ioutil.WriteFile(path.Join(s.installDir, s.pidfile), []byte(strconv.Itoa(cm.Process.Pid)), 0644)

}

func (s *startCMD) stop() error {
	strb, _ := ioutil.ReadFile(path.Join(s.installDir, s.pidfile))
	command := exec.Command("kill", string(strb))
	err := command.Start()
	if err != nil {
		return err
	}
	return command.Wait()
}

func (s *startCMD) runServer() error {

	if err := s.init(); err != nil {
		return err
	}
	installPath := path.Join(s.installDir, s.installName)
	err := s.execCMDPrintLog(installPath, "--config", path.Join(s.installDir, s.configName))
	if err != nil {
		return err
	}

	return nil
}

func (s *startCMD) execCMDPrintLog(name string, arg ...string) error {
	cm := exec.Command(name, arg...)
	stderr, _ := cm.StderrPipe()
	stdout, _ := cm.StdoutPipe()
	if err := cm.Start(); err != nil {
		return err
	}
	// 正常日志
	logScan := bufio.NewScanner(stdout)
	go func() {
		for logScan.Scan() {
			log.Println(logScan.Text())
		}
	}()
	// 错误日志
	errBuf := bytes.NewBufferString("")
	scan := bufio.NewScanner(stderr)
	for scan.Scan() {
		s := scan.Text()
		log.Println("build error: ", s)
		errBuf.WriteString(s)
		errBuf.WriteString("\n")
	}
	// 等待命令执行完
	cm.Wait()
	if !cm.ProcessState.Success() {
		// 执行失败，返回错误信息
		return errors.New(errBuf.String())
	}
	return nil
}

func (s *startCMD) downloadIfNeed() error {
	installPath := path.Join(s.installDir, s.installName)
	configPath := path.Join(s.installDir, s.configName)
	if !s.binaryIsExist() {
		tmpPath, err := s.downloadBinary()
		if err != nil {
			return err
		}
		err = s.mv(tmpPath, installPath)
		if err != nil {
			return err
		}
	}
	if !s.configIsExist() {
		tmpPath, err := s.downloadConfig()
		if err != nil {
			return err
		}
		err = s.mv(tmpPath, configPath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *startCMD) mv(oldPath, newPath string) error {
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

func (s *startCMD) binaryIsExist() bool {
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

func (s *startCMD) configIsExist() bool {
	configPath := path.Join(s.installDir, s.configName)
	_, err := os.Stat(configPath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func (s *startCMD) downloadBinary() (string, error) {

	downloadURL := s.downloadUrl
	if s.china {
		downloadURL = s.chinaDownloadUrl
	}

	downloadURL = strings.ReplaceAll(downloadURL, "${version}", s.version)
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
	_, err = io.Copy(writer, barReader)

	return destPath, err

}

func (s *startCMD) downloadConfig() (string, error) {
	downloadURL := s.configDowloadUrl
	if s.china {
		downloadURL = s.chinaConfigDowloadUrl
	}
	downloadURL = strings.ReplaceAll(downloadURL, "${version}", s.version)
	fmt.Println("Start download wukongim config from " + downloadURL + " ...")
	destPath := path.Join(os.TempDir(), "wukongim_config_tmp")

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
	_, err = io.Copy(writer, barReader)

	return destPath, err
}
