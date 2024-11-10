package cmd

type Channel struct {
	ChannelId   string
	ChannelType uint8
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

type ChannelInfoReq struct {
	ChannelId   string `json:"channel_id"`   // 频道ID
	ChannelType uint8  `json:"channel_type"` // 频道类型
	Large       int    `json:"large"`        // 是否是超大群
	Ban         int    `json:"ban"`          // 是否封禁频道（封禁后此频道所有人都将不能发消息，除了系统账号）
}

// ChannelCreateReq 频道创建请求
type ChannelCreateReq struct {
	ChannelInfoReq
	Subscribers []string `json:"subscribers"` // 订阅者
}

type SubscriberAddReq struct {
	ChannelId      string   `json:"channel_id"`      // 频道ID
	ChannelType    uint8    `json:"channel_type"`    // 频道类型
	Reset          int      `json:"reset"`           // 是否重置订阅者 （0.不重置 1.重置），选择重置，将删除原来的所有成员
	TempSubscriber int      `json:"temp_subscriber"` //  是否是临时订阅者 (1. 是 0. 否)
	Subscribers    []string `json:"subscribers"`     // 订阅者
}

type SubscriberReq struct {
	ChannelId   string   `json:"channel_id"`   // 频道ID
	ChannelType uint8    `json:"channel_type"` // 频道类型
	Subscribers []string `json:"subscribers"`  // 订阅者
}

type ChannelUidsReq struct {
	ChannelId   string   `json:"channel_id"`   // 频道ID
	ChannelType uint8    `json:"channel_type"` // 频道类型
	Uids        []string `json:"uids"`         // 用户ID集合
}
