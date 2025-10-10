package utils

import (
	"os/exec"
	"runtime"

	"github.com/cockroachdb/errors"
)

func OpenBrowser(url string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	case "linux":
		return exec.Command("xdg-open", url).Start()
	default:
		return errors.Newf("unsupported platform: %s", runtime.GOOS)
	}
}
