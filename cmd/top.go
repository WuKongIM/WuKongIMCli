package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gizak/termui/v3"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/spf13/cobra"
)

type Stats struct {
	Varz  *Varz
	Connz *Connz
	Rates *Rates
	Error error
}

type topCMD struct {
	ctx          *WuKongIMContext
	urlStr       string
	httpClient   *http.Client
	limit        int
	sortOpt      SortOpt
	delay        int
	statsCh      chan *Stats
	shutdownCh   chan struct{}
	lastStats    *Stats
	lastPollTime time.Time

	displayRawBytes bool
	displayClients  bool // 是否展示客户端

	maxStatsRefreshes int // 最大刷新次数 -1 表示无限制
	body              *ui.Grid
}

func newTopCMD(ctx *WuKongIMContext) *topCMD {
	b := &topCMD{
		ctx:               ctx,
		urlStr:            "http://127.0.0.1:1516",
		httpClient:        &http.Client{},
		delay:             1,
		statsCh:           make(chan *Stats),
		shutdownCh:        make(chan struct{}),
		maxStatsRefreshes: -1,
	}
	return b
}

func (t *topCMD) CMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "top",
		Short: "Data monitoring",
		RunE:  t.run,
	}
	t.initVar(cmd)
	return cmd
}

func (t *topCMD) initVar(cmd *cobra.Command) {

}

func (t *topCMD) run(cmd *cobra.Command, args []string) error {

	// Smoke test to abort in case can't connect to server since the beginning.
	_, err := t.request("/varz")
	if err != nil {
		return err
	}

	err = ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	go t.monitorStats()

	t.StartUI()

	return nil
}

func (t *topCMD) request(path string) (interface{}, error) {
	var statz interface{}
	uri := t.urlStr + path
	switch path {
	case "/varz":
		statz = &Varz{}
	case "/connz":
		statz = &Connz{}
		uri += fmt.Sprintf("?limit=%d&sort=%s", t.limit, t.sortOpt)
	default:
		return nil, fmt.Errorf("invalid path '%s' for stats server", path)
	}
	resp, err := t.httpClient.Get(uri)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("could not get stats from server: %w", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("stats request failed %d: %q", resp.StatusCode, string(body))
	}
	err = json.Unmarshal(body, &statz)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal statz json: %w", err)
	}

	return statz, nil
}

func (t *topCMD) monitorStats() error {
	delay := time.Duration(t.delay) * time.Second
	for {
		select {
		case <-t.shutdownCh:
			return nil
		case <-time.After(delay):
			t.statsCh <- t.fetchStats()
		}
	}
}

var errDud = fmt.Errorf("")

func (t *topCMD) fetchStats() *Stats {
	var inMsgsDelta int64
	var outMsgsDelta int64
	var inBytesDelta int64
	var outBytesDelta int64

	var inMsgsLastVal int64
	var outMsgsLastVal int64
	var inBytesLastVal int64
	var outBytesLastVal int64

	var inMsgsRate float64
	var outMsgsRate float64
	var inBytesRate float64
	var outBytesRate float64

	stats := &Stats{
		Varz:  &Varz{},
		Connz: &Connz{},
		Rates: &Rates{},
		Error: errDud,
	}
	// Get /varz
	{
		result, err := t.request("/varz")
		if err != nil {
			stats.Error = err
			return stats
		}

		if varz, ok := result.(*Varz); ok {
			stats.Varz = varz
		}
	}
	// Get /connz
	{
		result, err := t.request("/connz")
		if err != nil {
			stats.Error = err
			return stats
		}

		if connz, ok := result.(*Connz); ok {
			stats.Connz = connz
		}
	}

	var isFirstTime bool
	if t.lastStats != nil {
		inMsgsLastVal = t.lastStats.Varz.InMsgs
		outMsgsLastVal = t.lastStats.Varz.OutMsgs
		inBytesLastVal = t.lastStats.Varz.InBytes
		outBytesLastVal = t.lastStats.Varz.OutBytes
	} else {
		isFirstTime = true
	}

	// Periodic snapshot to get per sec metrics
	inMsgsVal := stats.Varz.InMsgs
	outMsgsVal := stats.Varz.OutMsgs
	inBytesVal := stats.Varz.InBytes
	outBytesVal := stats.Varz.OutBytes

	inMsgsDelta = inMsgsVal - inMsgsLastVal
	outMsgsDelta = outMsgsVal - outMsgsLastVal
	inBytesDelta = inBytesVal - inBytesLastVal
	outBytesDelta = outBytesVal - outBytesLastVal

	inMsgsLastVal = inMsgsVal
	outMsgsLastVal = outMsgsVal
	inBytesLastVal = inBytesVal
	outBytesLastVal = outBytesVal

	now := time.Now()
	tdelta := now.Sub(t.lastPollTime)

	// Calculate rates but the first time
	if !isFirstTime {
		inMsgsRate = float64(inMsgsDelta) / tdelta.Seconds()
		outMsgsRate = float64(outMsgsDelta) / tdelta.Seconds()
		inBytesRate = float64(inBytesDelta) / tdelta.Seconds()
		outBytesRate = float64(outBytesDelta) / tdelta.Seconds()
	}
	rates := &Rates{
		InMsgsRate:   inMsgsRate,
		OutMsgsRate:  outMsgsRate,
		InBytesRate:  inBytesRate,
		OutBytesRate: outBytesRate,
	}
	stats.Rates = rates

	// Snapshot stats.
	t.lastStats = stats
	t.lastPollTime = now

	return stats
}

func (t *topCMD) StartUI() {
	cleanStats := &Stats{
		Varz:  &Varz{},
		Connz: &Connz{},
		Rates: &Rates{},
		Error: fmt.Errorf(""),
	}
	// Show empty values on first display

	termWidth, termHeight := ui.TerminalDimensions()

	text := t.generateParagraph(cleanStats, "")

	body := widgets.NewTable()
	body.SetRect(0, 0, termWidth, termHeight)

	par := widgets.NewParagraph()
	par.Text = text
	par.Border = false

	par.SetRect(0, 0, termWidth, termHeight)

	helpText := generateHelp()
	helpPar := widgets.NewParagraph()
	helpPar.Text = helpText
	helpPar.Border = false
	helpPar.SetRect(0, 0, termWidth, termHeight)

	// Top like view
	topViewGrid := ui.NewGrid()
	topViewGrid.SetRect(0, 0, termWidth, termHeight)
	topViewGrid.Set(ui.NewRow(1, ui.NewCol(1, par)))

	// Help view
	helpViewGrid := ui.NewGrid()
	helpViewGrid.SetRect(0, 0, termWidth, termHeight)
	helpViewGrid.Set(ui.NewRow(1, ui.NewCol(1, helpPar)))

	// Used to toggle back to previous mode
	viewMode := TopViewMode

	// Used for pinging the IU to refresh the screen with new values
	redraw := make(chan RedrawCause)

	update := func() {
		for {
			stats := <-t.statsCh

			par.Text = t.generateParagraph(stats, "") // Update top view text

			redraw <- DueToNewStats
		}
	}

	// Flags for capturing options
	waitingSortOption := false
	waitingLimitOption := false

	optionBuf := ""
	refreshOptionHeader := func() {
		clrline := "\033[1;1H\033[6;1H                  " // Need to mask what was typed before

		clrline += "  "
		for i := 0; i < len(optionBuf); i++ {
			clrline += "  "
		}
		fmt.Print(clrline)
	}

	t.body = topViewGrid

	ui.Render(t.body)

	go update()

	numberOfRedrawsDueToNewStats := 0

	evt := ui.PollEvents()

	for {
		select {
		case e := <-evt:

			if waitingSortOption {

				if e.Type == ui.KeyboardEvent && e.ID == "<Enter>" {
					sortOpt := SortOpt(optionBuf)
					if sortOpt.IsValid() {
						t.sortOpt = sortOpt
					} else {
						go func() {
							// Has to be at least of the same length as sort by header
							emptyPadding := "       "
							fmt.Printf("\033[1;1H\033[6;1Hinvalid order: %s%s", optionBuf, emptyPadding)
							waitingSortOption = false
							time.Sleep(1 * time.Second)
							refreshOptionHeader()
							optionBuf = ""
						}()
						continue
					}
					refreshOptionHeader()
					waitingSortOption = false
					optionBuf = ""
					continue
				}
				// Handle backspace
				if e.Type == ui.KeyboardEvent && len(optionBuf) > 0 && (e.ID == "<C-<Backspace>>" || e.ID == "<Backspace>") {
					optionBuf = optionBuf[:len(optionBuf)-1]
					refreshOptionHeader()
				} else {
					optionBuf += e.ID
				}
				fmt.Printf("\033[1;1H\033[6;1Hsort by [%s]: %s", t.sortOpt, optionBuf)

			}
			if waitingLimitOption {
				if e.Type == ui.KeyboardEvent && e.ID == "<Enter>" {
					var n int
					_, err := fmt.Sscanf(optionBuf, "%d", &n)
					if err == nil {
						t.limit = n
					}

					waitingLimitOption = false
					optionBuf = ""
					refreshOptionHeader()
					continue
				}
				// Handle backspace
				if e.Type == ui.KeyboardEvent && len(optionBuf) > 0 && (e.ID == "<C-<Backspace>>" || e.ID == "<Backspace>") {
					optionBuf = optionBuf[:len(optionBuf)-1]
					refreshOptionHeader()
				} else {
					optionBuf += e.ID
				}
				fmt.Printf("\033[1;1H\033[6;1Hlimit   [%d]: %s", t.limit, optionBuf)
			}
			if e.Type == ui.KeyboardEvent && (e.ID == "q" || e.ID == "<C-c>") {
				close(t.shutdownCh)
				cleanExit()
			}

			if e.Type == ui.KeyboardEvent && e.ID == "o" && !waitingLimitOption && viewMode == TopViewMode {
				fmt.Printf("\033[1;1H\033[6;1Hsort by [%s]:", t.sortOpt)
				waitingSortOption = true
			}

			if e.Type == ui.KeyboardEvent && e.ID == "n" && !waitingSortOption && viewMode == TopViewMode {
				fmt.Printf("\033[1;1H\033[6;1Hlimit   [%d]:", t.limit)
				waitingLimitOption = true
			}
			if e.Type == ui.KeyboardEvent && (e.ID == "?" || e.ID == "h") && !(waitingSortOption || waitingLimitOption) {
				if viewMode == TopViewMode {
					refreshOptionHeader()
					optionBuf = ""
				}

				t.body = helpViewGrid
				viewMode = HelpViewMode
				waitingLimitOption = false
				waitingSortOption = false
			}

			if e.Type == ui.KeyboardEvent && (e.ID == "b") && !(waitingSortOption || waitingLimitOption) {
				t.displayRawBytes = !t.displayRawBytes
			}

			if e.Type == termui.ResizeEvent {
				w, h := termui.TerminalDimensions()
				t.body.SetRect(0, 0, w, h)
				go func() { redraw <- DueToViewportResize }()
			}

		case cause := <-redraw:
			ui.Render(t.body)

			if cause == DueToNewStats {
				numberOfRedrawsDueToNewStats += 1

				if t.maxStatsRefreshes > 0 && numberOfRedrawsDueToNewStats >= t.maxStatsRefreshes {
					close(t.shutdownCh)
					cleanExit()
				}
			}
		}
	}
}

func (t *topCMD) generateParagraph(
	stats *Stats,
	outputDelimiter string,
) string {

	return t.generateParagraphPlainText(stats)
}

func (t *topCMD) generateParagraphPlainText(
	stats *Stats,
) string {
	// Snapshot current stats
	cpu := stats.Varz.CPU
	memVal := stats.Varz.Mem
	uptime := stats.Varz.Uptime
	numConns := stats.Connz.Total
	inMsgsVal := stats.Varz.InMsgs
	outMsgsVal := stats.Varz.OutMsgs
	inBytesVal := stats.Varz.InBytes
	outBytesVal := stats.Varz.OutBytes
	slowConsumers := stats.Varz.SlowConsumers

	var serverVersion string
	if stats.Varz.Version != "" {
		serverVersion = stats.Varz.Version
	}

	mem := Psize(false, memVal) //memory is exempt from the rawbytes flag
	inMsgs := Psize(t.displayRawBytes, inMsgsVal)
	outMsgs := Psize(t.displayRawBytes, outMsgsVal)
	inBytes := Psize(t.displayRawBytes, inBytesVal)
	outBytes := Psize(t.displayRawBytes, outBytesVal)
	inMsgsRate := stats.Rates.InMsgsRate
	outMsgsRate := stats.Rates.OutMsgsRate
	inBytesRate := Psize(t.displayRawBytes, int64(stats.Rates.InBytesRate))
	outBytesRate := Psize(t.displayRawBytes, int64(stats.Rates.OutBytesRate))

	info := "WuKongIM server version %s (uptime: %s) %s\n"
	info += "Server:\n"
	info += "  Load: CPU:  %.1f%%  Memory: %s  Slow Consumers: %d\n"
	info += "  In:   Msgs: %s  Bytes: %s  Msgs/Sec: %.1f  Bytes/Sec: %s\n"
	info += "  Out:  Msgs: %s  Bytes: %s  Msgs/Sec: %.1f  Bytes/Sec: %s"

	text := fmt.Sprintf(
		info, serverVersion, uptime, stats.Error,
		cpu, mem, slowConsumers,
		inMsgs, inBytes, inMsgsRate, inBytesRate,
		outMsgs, outBytes, outMsgsRate, outBytesRate,
	)

	text += fmt.Sprintf("\n\nConnections Polled: %d\n", numConns)

	displayClients := t.displayClients

	header := make([]interface{}, 0) // Dynamically add columns and padding depending
	hostSize := DEFAULT_HOST_PADDING_SIZE

	uidSize := 0 // Disable uid unless we have seen one using it

	for _, conn := range stats.Connz.Connections {
		var size int

		var hostname = fmt.Sprintf("%s:%d", conn.IP, conn.Port)

		size = len(hostname) // host
		if size > hostSize {
			hostSize = size + DEFAULT_PADDING_SIZE
		}

		size = len(conn.UID) // name
		if size > uidSize {
			uidSize = size + DEFAULT_PADDING_SIZE

			minLen := len("UID") // If using name, ensure that it is not too small...
			if uidSize < minLen {
				uidSize = minLen
			}
		}
	}

	connHeader := DEFAULT_PADDING // Initial padding

	header = append(header, "HOST") // HOST
	connHeader += "%-" + fmt.Sprintf("%d", hostSize) + "s "

	header = append(header, "ID") // ID
	connHeader += " %-6s "

	if uidSize > 0 { // uid
		header = append(header, "UID")
		connHeader += "%-" + fmt.Sprintf("%d", uidSize) + "s "
	}

	header = append(header, standardHeaders...)

	connHeader += strings.Join(defaultHeaderColumns, "  ")
	if displayClients {
		connHeader += "%13s"
	}

	connHeader += "\n" // ...LAST ACTIVITY

	var connRows string
	if displayClients {
		header = append(header, "SUBSCRIPTIONS")
	}

	connRows = fmt.Sprintf(connHeader, header...)

	text += connRows // Add to screen!

	connValues := DEFAULT_PADDING

	connValues += "%-" + fmt.Sprintf("%d", hostSize) + "s " // HOST: e.g. 192.168.1.1:78901

	connValues += " %-6d " // CID: e.g. 1234

	if uidSize > 0 { // uid: e.g. uid1000
		connValues += "%-" + fmt.Sprintf("%d", uidSize) + "s "
	}

	connValues += strings.Join(defaultRowColumns, "  ")
	if displayClients {
		connValues += "%s"
	}
	connValues += "\n"

	for _, conn := range stats.Connz.Connections {
		var h = fmt.Sprintf("%s:%d", conn.IP, conn.Port)

		var connLine string // Build the info line
		connLineInfo := make([]interface{}, 0)
		connLineInfo = append(connLineInfo, h)
		connLineInfo = append(connLineInfo, conn.ID)

		if uidSize > 0 { // uid not included unless present
			connLineInfo = append(connLineInfo, conn.UID)
		}

		connLineInfo = append(connLineInfo, 0)
		connLineInfo = append(connLineInfo, Psize(t.displayRawBytes, int64(conn.PendingBytes)), Psize(t.displayRawBytes, conn.OutMsgs), Psize(t.displayRawBytes, conn.InMsgs))
		connLineInfo = append(connLineInfo, Psize(t.displayRawBytes, conn.OutBytes), Psize(t.displayRawBytes, conn.InBytes))

		deviceID := conn.DeviceID
		if len(deviceID) > 13 {
			deviceID = fmt.Sprintf("%s-%s", deviceID[:6], deviceID[len(deviceID)-6:])
		}
		connLineInfo = append(connLineInfo, deviceID)                        // 设备ID
		connLineInfo = append(connLineInfo, conn.Device)                     // 设备
		connLineInfo = append(connLineInfo, fmt.Sprintf("%d", conn.Version)) // 客户端协议版本
		connLineInfo = append(connLineInfo, conn.Uptime, conn.LastActivity)

		if t.displayClients {
			connLineInfo = append(connLineInfo, "subs")
		}

		connLine = fmt.Sprintf(connValues, connLineInfo...)

		text += connLine // Add line to screen!
	}

	return text
}

type Connz struct {
	Connections []*ConnInfo `json:"connections"` // 连接数
	Now         time.Time   `json:"now"`         // 查询时间
	Total       int         `json:"total"`       // 总连接数量
	Offset      int         `json:"offset"`      // 偏移位置
	Limit       int         `json:"limit"`       // 限制数量
}

type ConnInfo struct {
	ID           int64     `json:"id"`            // 连接ID
	UID          string    `json:"uid"`           // 用户uid
	IP           string    `json:"ip"`            // 客户端IP
	Port         int       `json:"port"`          // 客户端端口
	LastActivity time.Time `json:"last_activity"` // 最后一次活动时间
	Uptime       string    `json:"uptime"`        // 启动时间
	Idle         string    `json:"idle"`          // 客户端闲置时间
	PendingBytes int       `json:"pending_bytes"` // 等待发送的字节数
	InMsgs       int64     `json:"in_msgs"`       // 流入的消息数
	OutMsgs      int64     `json:"out_msgs"`      // 流出的消息数量
	InBytes      int64     `json:"in_bytes"`      // 流入的字节数量
	OutBytes     int64     `json:"out_bytes"`     // 流出的字节数量
	Device       string    `json:"device"`        // 设备
	DeviceID     string    `json:"device_id"`     // 设备ID
	Version      uint8     `json:"version"`       // 客户端协议版本
}

type Rates struct {
	InMsgsRate   float64
	OutMsgsRate  float64
	InBytesRate  float64
	OutBytesRate float64
}

type SortOpt string

const (
	ByID SortOpt = "id" // By connection ID
)

func (s SortOpt) IsValid() bool {
	switch s {
	case "", ByID:
		return true
	default:
		return false
	}
}

const kibibyte = 1024
const mebibyte = 1024 * 1024
const gibibyte = 1024 * 1024 * 1024

// Psize takes a float and returns a human readable string.
func Psize(displayRawValue bool, s int64) string {
	size := float64(s)

	if displayRawValue || size < kibibyte {
		return fmt.Sprintf("%.0f", size)
	}

	if size < mebibyte {
		return fmt.Sprintf("%.1fK", size/kibibyte)
	}

	if size < gibibyte {
		return fmt.Sprintf("%.1fM", size/mebibyte)
	}

	return fmt.Sprintf("%.1fG", size/gibibyte)
}

const (
	DEFAULT_PADDING_SIZE = 2
	DEFAULT_PADDING      = "  "

	DEFAULT_HOST_PADDING_SIZE = 15
)

var (
	resolvedHosts = map[string]string{} // cache for reducing DNS lookups in case enabled

	standardHeaders = []interface{}{"CLIENTS", "PENDING", "MSGS_TO", "MSGS_FROM", "BYTES_TO", "BYTES_FROM", "DEVICE_ID", "DEVICE", "VERSION", "UPTIME", "LAST_ACTIVITY"}

	defaultHeaderColumns = []string{"%-6s", "%-10s", "%-10s", "%-10s", "%-10s", "%-10s", "%-10s", "%-7s", "%-7s", "%-7s", "%-40s"} // Chopped: HOST ID UID...
	defaultRowColumns    = []string{"%-6d", "%-10s", "%-10s", "%-10s", "%-10s", "%-10s", "%-10s", "%-7s", "%-7s", "%-7s", "%-40s"}
)

type ViewMode int

const (
	TopViewMode ViewMode = iota
	HelpViewMode
)

type RedrawCause int

const (
	DueToNewStats RedrawCause = iota
	DueToViewportResize
)

func cleanExit() {
	clearScreen()
	ui.Close()

	// Show cursor once again
	fmt.Print("\033[?25h")
	os.Exit(0)
}

// clearScreen tries to ensure resetting original state of screen
func clearScreen() {
	fmt.Print("\033[2J\033[1;1H\033[?25l")
}

func generateHelp() string {
	text := `
Command          Description

o<option>        Set primary sort key to <option>.

                 Option can be one of: {cid|subs|pending|msgs_to|msgs_from|
                 bytes_to|bytes_from|idle|last}

                 This can be set in the command line too with -sort flag.

n<limit>         Set sample size of connections to request from the server.

                 This can be set in the command line as well via -n flag.
                 Note that if used in conjunction with sort, the server
                 would respect both options allowing queries like 'connection
                 with largest number of subscriptions': -n 1 -sort subs

s                Toggle displaying connection subscriptions.

d                Toggle activating DNS address lookup for clients.

b                Toggle displaying raw bytes.

q                Quit nats-top.

Press any key to continue...

`
	return text
}
