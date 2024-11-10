package cmd

import (
	"errors"
	"net/http"

	"github.com/WuKongIM/WuKongIMCli/pkg/network"
	"github.com/WuKongIM/WuKongIMCli/pkg/wkutil"
	wkproto "github.com/WuKongIM/WuKongIMGoProto"
	"github.com/sendgrid/rest"
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

func (a *API) UpdateToken(uid, token string, deviceFlag wkproto.DeviceFlag, deviceLevel wkproto.DeviceLevel) error {

	resp, err := network.Post(a.getFullURL("/user/token"), []byte(wkutil.ToJSON(map[string]interface{}{"uid": uid, "token": token, "device_flag": deviceFlag, "device_level": deviceLevel})), nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return a.handleError(resp)
	}
	return nil
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

func (a *API) CreateChannel(req *ChannelCreateReq) error {

	resp, err := network.Post(a.getFullURL("/channel"), []byte(wkutil.ToJSON(req)), nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return a.handleError(resp)
	}
	return nil
}

// SubscriberAdd 添加订阅者
func (a *API) SubscriberAdd(req *SubscriberAddReq) error {
	resp, err := network.Post(a.getFullURL("/channel/subscriber_add"), []byte(wkutil.ToJSON(req)), nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return a.handleError(resp)
	}
	return nil
}

// SubscriberRemove 移除订阅者
func (a *API) SubscriberRemove(req *SubscriberReq) error {
	resp, err := network.Post(a.getFullURL("/channel/subscriber_remove"), []byte(wkutil.ToJSON(req)), nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return a.handleError(resp)
	}
	return nil
}

// DenylistAdd 添加黑名单
func (a *API) DenylistAdd(req *ChannelUidsReq) error {
	resp, err := network.Post(a.getFullURL("/channel/blacklist_add"), []byte(wkutil.ToJSON(req)), nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return a.handleError(resp)
	}
	return nil
}

// DenylistRemove 移除黑名单
func (a *API) DenylistRemove(req *ChannelUidsReq) error {
	resp, err := network.Post(a.getFullURL("/channel/blacklist_remove"), []byte(wkutil.ToJSON(req)), nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return a.handleError(resp)
	}
	return nil
}

func (a *API) AllowlistAdd(req *ChannelUidsReq) error {
	resp, err := network.Post(a.getFullURL("/channel/whitelist_add"), []byte(wkutil.ToJSON(req)), nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return a.handleError(resp)
	}
	return nil
}

func (a *API) AllowlistRemove(req *ChannelUidsReq) error {
	resp, err := network.Post(a.getFullURL("/channel/whitelist_remove"), []byte(wkutil.ToJSON(req)), nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return a.handleError(resp)
	}
	return nil
}

func (a *API) handleError(resp *rest.Response) error {
	if resp.StatusCode == http.StatusBadRequest {
		resultMap, err := wkutil.JSONToMap(resp.Body)
		if err != nil {
			return err
		}
		msg, ok := resultMap["msg"]
		if ok {
			return errors.New(msg.(string))
		}
	}
	return errors.New("未知错误")
}

func (a *API) getFullURL(path string) string {
	return a.baseURL + path
}
