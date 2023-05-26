package cmd

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

type doctorCMD struct {
}

func newDoctorCMD(ctx *WuKongIMContext) *doctorCMD {
	return &doctorCMD{}
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

	return nil
}

func ScanPort(protocol string, hostname string, port int) bool {
	fmt.Printf("scanning port %d \n", port)
	p := strconv.Itoa(port)
	addr := net.JoinHostPort(hostname, p)
	conn, err := net.DialTimeout(protocol, addr, 3*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
