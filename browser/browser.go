package browser

import (
	"fmt"
	"os/exec"
	"runtime"
)

// Open browser for different os.
func Open(url string) (*exec.Cmd, error) {
	return commandForOS(runtime.GOOS, url)
}

func commandForOS(goos, url string) (*exec.Cmd, error) {
	var exe string
	var args []string

	switch goos {
	case "linux":
		exe = "xdg-open"
		args = append(args, url)
	case "windows":
		exe = "rundll32"
		args = append(args, "url.dll,FileProtocolHandler", url)
	case "darwin":
		exe = "open"
		args = append(args, url)
	default:
		return nil, fmt.Errorf("unsupported platform %s", goos)
	}

	return exec.Command(exe, args...), nil
}
