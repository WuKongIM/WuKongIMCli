package limutil

import (
	"os/exec"
	"strings"
)

func ExecCMD(aCMD string, arg ...string) (string, error) {
	var cmdText string
	out, err := exec.Command(aCMD, arg...).Output()
	if err != nil {
		return "", err
	}
	cmdText = string(out)
	return strings.TrimSpace(cmdText), nil
}
