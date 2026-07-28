//go:build !(aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris)

package workflow

func staleLease(string) (bool, error) {
	return false, nil
}
