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

func (a *API) getFullURL(path string) string {
	return a.baseURL + path
}

type userAddrResp struct {
	TCPAddr string   `json:"tcp_addr"`
	WSAddr  string   `json:"ws_addr"`
	UIDs    []string `json:"uids"`
}
