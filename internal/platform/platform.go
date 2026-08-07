package platform

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func PickFileOrFolder(prompt string) (string, error) {
	if runtime.GOOS != "darwin" {
		return "", errors.New("native picker is currently available on macOS only")
	}
	script := fmt.Sprintf(`set chosenItem to choose file with prompt %q
POSIX path of chosenItem`, prompt)
	out, err := exec.Command("osascript", "-e", script).Output()
	if err == nil {
		return strings.TrimSpace(string(out)), nil
	}
	script = fmt.Sprintf(`set chosenItem to choose folder with prompt %q
POSIX path of chosenItem`, prompt)
	out, err = exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func PickFolder(prompt string) (string, error) {
	if runtime.GOOS != "darwin" {
		return "", errors.New("native folder picker is currently available on macOS only")
	}
	script := fmt.Sprintf(`set chosenItem to choose folder with prompt %q
POSIX path of chosenItem`, prompt)
	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func Notify(title, message string) error {
	if runtime.GOOS != "darwin" {
		return nil
	}
	script := fmt.Sprintf(`display notification %q with title %q`, message, title)
	return exec.Command("osascript", "-e", script).Run()
}

func OpenURL(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("URL is empty")
	}
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", value).Run()
	case "linux":
		return exec.Command("xdg-open", value).Run()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", value).Run()
	default:
		return errors.New("opening URLs is not supported on this platform")
	}
}

func OpenFolder(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return errors.New("folder path is empty")
	}
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", path).Run()
	case "linux":
		return exec.Command("xdg-open", path).Run()
	case "windows":
		return exec.Command("explorer", path).Run()
	default:
		return errors.New("opening folders is not supported on this platform")
	}
}
