package cmd

import (
	"errors"
	"net/http"

	"github.com/WuKongIM/WuKongIMCli/pkg/network"
	"github.com/WuKongIM/WuKongIMCli/pkg/wkutil"
)

type API struct {
	baseURL string
}

func NewAPI() *API {
	return &API{}
}

func (a *API) SetBaseURL(baseURL string) {
	a.baseURL = baseURL
}

func (a *API) Route(uids []string) (map[string]string, error) {

	resp, err := network.Post(a.getFullURL("/route/batch"), []byte(wkutil.ToJSON(uids)), nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("请求失败")
	}

	var userAddrs []userAddrResp
	err = wkutil.ReadJSONByByte([]byte(resp.Body), &userAddrs)
	if err != nil {
		return nil, err
	}

	resultMap := make(map[string]string)
	if len(userAddrs) > 0 {
		for _, userAddr := range userAddrs {
			if len(userAddr.UIDs) > 0 {
				for _, uid := range userAddr.UIDs {
					resultMap[uid] = userAddr.TCPAddr
				}
			}
		}
	}
	return resultMap, nil

}

func (a *API) Varz() (*Varz, error) {
	resp, err := network.Get(a.getFullURL("/varz"), nil, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("请求失败")
	}
	var varz *Varz
	err = wkutil.ReadJSONByByte([]byte(resp.Body), &varz)
	if err != nil {
		return nil, err
	}
	return varz, nil
}

func (a *API) getFullURL(path string) string {
	return a.baseURL + path
}

type userAddrResp struct {
	TCPAddr string   `json:"tcp_addr"`
	WSAddr  string   `json:"ws_addr"`
	UIDs    []string `json:"uids"`
}

type Varz struct {
	ServerID    string  `json:"server_id"`   // 服务端ID
	ServerName  string  `json:"server_name"` // 服务端名称
	Version     string  `json:"version"`     // 服务端版本
	Connections int     `json:"connections"` // 当前连接数量
	Uptime      string  `json:"uptime"`      // 上线时间
	Mem         int64   `json:"mem"`         // 内存
	CPU         float64 `json:"cpu"`         // cpu

	InMsgs        int64 `json:"in_msgs"`        // 流入消息数量
	OutMsgs       int64 `json:"out_msgs"`       // 流出消息数量
	InBytes       int64 `json:"in_bytes"`       // 流入字节数量
	OutBytes      int64 `json:"out_bytes"`      // 流出字节数量
	SlowConsumers int64 `json:"slow_consumers"` // 慢客户端数量

	TCPAddr     string `json:"tcp_addr"`     // tcp地址
	WSAddr      string `json:"ws_addr"`      // wss地址
	MonitorAddr string `json:"monitor_addr"` // 监控地址
	MonitorOn   int    `json:"monitor_on"`   // 监控是否开启
	Commit      string `json:"commit"`       // git commit id
	CommitDate  string `json:"commit_date"`  // git commit date
	TreeState   string `json:"tree_state"`   // git tree state
}
