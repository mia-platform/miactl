package browser

import (
	"fmt"
	"os/exec"
)

func OpenBrowser(goos, url string) error {
	var err error

	switch goos {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform %s", goos)
	}

	return err
}
