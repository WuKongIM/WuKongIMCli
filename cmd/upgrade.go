package cmd

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/WuKongIM/WuKongIMCli/pkg/wkutil"
	"github.com/cheggaaa/pb/v3"
	"github.com/spf13/cobra"
)

type upgradeCMD struct {
	installDir  string
	downloadUrl string
	installName string
}

func newUpgradeCMD(ctx *WuKongIMContext) *upgradeCMD {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return &upgradeCMD{
		installDir:  path.Join(homeDir, "wukongim"),
		installName: "wukongim",
		downloadUrl: "https://gitee.com/WuKongDev/WuKongIM/releases/download/${version}/wukongim-${sysos}-${sysarch}",
	}
}

func (u *upgradeCMD) CMD() *cobra.Command {
	upgradeCmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade WuKong IM service",
		RunE: func(cmd *cobra.Command, args []string) error {
			return u.upgrade()
		},
	}
	return upgradeCmd
}

func (u *upgradeCMD) upgrade() error {

	versionBytes, err := ioutil.ReadFile(path.Join(u.installDir, "version"))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	version := ""
	if len(versionBytes) > 0 {
		version = string(versionBytes)
	}

	// 获取最佳版本
	fmt.Printf("Get the latest version \n")
	lastVersion, err := u.getLastVersion()
	if err != nil {
		return err
	}
	if strings.TrimSpace(version) == strings.TrimSpace(lastVersion) { // 版本一致，不需要升级
		fmt.Printf("Current version %s is already up to date. \n", version)
		return nil
	}
	fmt.Printf("Current version is %s, the latest version is %s. Let's start the upgrade. \n", version, lastVersion)
	tmpath, err := u.downloadBinary(lastVersion)
	if err != nil {
		return err
	}
	installPath := path.Join(u.installDir, u.installName)
	err = move(tmpath, installPath)
	if err != nil {
		return err
	}
	err = u.writeVersion(lastVersion)
	if err != nil {
		return err
	}
	fmt.Printf("Upgrade to version %s successfully, please restart for the changes to take effect. \n", lastVersion)
	return nil
}

func (u *upgradeCMD) writeVersion(version string) error {
	versionFile := path.Join(u.installDir, "version")
	return ioutil.WriteFile(versionFile, []byte(version), 0644)
}

func (u *upgradeCMD) getLastVersion() (string, error) {
	releaseUrl := "https://gitee.com/api/v5/repos/WuKongDev/WuKongIM/releases/latest"
	resp, err := http.Get(releaseUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("get latest version failed")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	releaseResultMap, err := wkutil.JSONToMap(string(body))
	if err != nil {
		return "", err
	}
	lastVersion := releaseResultMap["tag_name"].(string)
	return lastVersion, nil

}

func (u *upgradeCMD) downloadBinary(version string) (string, error) {

	sysos := strings.ToLower(runtime.GOOS)
	sysarch := strings.ToLower(runtime.GOARCH)
	downloadURL := u.downloadUrl

	downloadURL = strings.ReplaceAll(downloadURL, "${version}", version)
	downloadURL = strings.ReplaceAll(downloadURL, "${sysos}", sysos)
	downloadURL = strings.ReplaceAll(downloadURL, "${sysarch}", sysarch)

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
