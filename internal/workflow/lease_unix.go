//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package workflow

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func staleLease(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	value := strings.TrimSpace(string(data))
	if !strings.HasPrefix(value, "pid=") {
		return false, fmt.Errorf("lease has invalid contents")
	}
	pid, err := strconv.Atoi(strings.TrimPrefix(value, "pid="))
	if err != nil || pid <= 0 {
		return false, fmt.Errorf("lease has invalid pid")
	}
	err = syscall.Kill(pid, 0)
	if err == nil || errors.Is(err, syscall.EPERM) {
		return false, nil
	}
	if errors.Is(err, syscall.ESRCH) {
		return true, nil
	}
	return false, err
}
