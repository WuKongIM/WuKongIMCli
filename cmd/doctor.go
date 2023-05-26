package cmd

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type doctorCMD struct {
	ctx *WuKongIMContext
	api *API
}

func newDoctorCMD(ctx *WuKongIMContext) *doctorCMD {
	return &doctorCMD{
		ctx: ctx,
		api: NewAPI(),
	}
}

func (d *doctorCMD) CMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check service status",
		RunE:  d.run,
	}
	return cmd
}

func (d *doctorCMD) run(cmd *cobra.Command, args []string) error {
	fmt.Println("Checking service status...")

	issueCount := 0

	// http server
	check, port, err := d.checkTCP(d.ctx.opts.ServerAddr)
	if err != nil {
		return err
	}
	if check {
		fmt.Printf("\x1B[32m%s %d\x1b[0m\n", "[✓] Http Service is running in", port)
	} else {
		fmt.Printf("\x1B[31m%s %d\x1b[0m\n", "[x] Http Service is not running in", port)
		issueCount++
	}
	if check {
		d.api.SetBaseURL(d.ctx.opts.ServerAddr)

		varz, err := d.api.Varz()
		if err != nil {
			return err
		}
		// tcp server
		check, port, err = d.checkTCP(varz.TCPAddr)
		if err != nil {
			return err
		}
		if check {
			fmt.Printf("\x1B[32m%s %d\x1b[0m\n", "[✓] TCP Service is running in", port)
		} else {
			fmt.Printf("\x1B[31m%s %d\x1b[0m\n", "[x] TCP Service is not running in", port)
			issueCount++
		}

		// websocket server
		// check, port, err = d.checkTCP(varz.WSSAddr)
		// if err != nil {
		// 	return err
		// }
		// if check {
		// 	fmt.Printf("Websocket Service is running in %d.\n", port)
		// } else {
		// 	fmt.Printf("Websocket Service is not running in %d.\n", port)
		// }

		// monitor server
		if varz.MonitorOn == 1 {
			check, port, err = d.checkTCP(varz.MonitorAddr)
			if err != nil {
				return err
			}
			if check {
				fmt.Printf("\x1B[32m%s %d\x1b[0m\n", "[✓] Monitor Service is running in", port)
			} else {
				fmt.Printf("\x1B[31m%s %d\x1b[0m\n", "[x] Monitor Service is not running in", port)
				issueCount++
			}
		}
	}
	if issueCount > 0 {
		fmt.Printf("\x1B[31m%s \x1b[0m\n", fmt.Sprintf("Found %d issues", issueCount))
	} else {
		fmt.Printf("\x1B[32m%s \x1b[0m\n", "• No issues found!")
	}

	return nil
}

func (d *doctorCMD) checkTCP(tcpAddr string) (bool, int, error) {
	var (
		port     int
		hostname string
	)
	if strings.Contains(tcpAddr, "://") {
		u, err := url.Parse(tcpAddr)
		if err != nil {
			return false, 0, err
		}
		port, _ = strconv.Atoi(u.Port())
		hostname = u.Hostname()
	} else {
		host, portStr, _ := net.SplitHostPort(tcpAddr)
		hostname = host
		port, _ = strconv.Atoi(portStr)
	}

	check := ScanPort("tcp", hostname, port)

	return check, port, nil
}

func ScanPort(protocol string, hostname string, port int) bool {
	p := strconv.Itoa(port)
	addr := net.JoinHostPort(hostname, p)
	conn, err := net.DialTimeout(protocol, addr, 3*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
