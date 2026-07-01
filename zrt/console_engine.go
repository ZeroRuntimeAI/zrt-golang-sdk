package zrt

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// requiredEngineVersion is the zrt-console engine channel the SDK expects.
const requiredEngineVersion = "latest"

// errConsoleRoomRequired is returned when console mode is started without a room.
var errConsoleRoomRequired = errors.New(
	"console mode requires a room id (set RoomOptions.RoomID, or provide auth so a room can be created)")

func detectTarget() (string, error) {
	var arch string
	switch runtime.GOARCH {
	case "amd64":
		arch = "x86_64"
	case "arm64":
		arch = "aarch64"
	default:
		return "", fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}
	switch runtime.GOOS {
	case "darwin":
		return arch + "-apple-darwin", nil
	case "linux":
		return arch + "-unknown-linux-gnu", nil
	case "windows":
		return arch + "-pc-windows-msvc", nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// engineBinaryName returns the platform-specific engine binary filename.
func engineBinaryName() string {
	if runtime.GOOS == "windows" {
		return "zrt-console.exe"
	}
	return "zrt-console"
}

// zrtHome returns the base directory for installed engine binaries.
func zrtHome() string {
	if h := strings.TrimSpace(os.Getenv("ZRT_HOME")); h != "" {
		return h
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, "zrt")
}

// engineInstallPath returns the canonical install path for the engine binary.
func engineInstallPath(version, target string) string {
	return filepath.Join(zrtHome(), "bin", "engine", version, target, engineBinaryName())
}

// isExec reports whether path is an existing regular executable file.
func isExec(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	// On Windows the executable bit is not meaningful; existence is enough.
	if runtime.GOOS == "windows" {
		return true
	}
	return info.Mode()&0o111 != 0
}

// localBuildFallback walks up from the working directory looking for a locally

func localBuildFallback() string {
	name := engineBinaryName()
	cur, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		for _, profile := range []string{"release", "debug"} {
			cand := filepath.Join(cur, "zrt", "target", profile, name)
			if isExec(cand) {
				return cand
			}
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return ""
		}
		cur = parent
	}
}

// resolveEngineBinary locates the zrt-console engine, downloading it if needed.
func resolveEngineBinary(version string) (string, error) {
	target, err := detectTarget()
	if err != nil {
		return "", err
	}
	installed := engineInstallPath(version, target)
	if isExec(installed) {
		return installed, nil
	}
	if override := strings.TrimSpace(os.Getenv("ZRT_CONSOLE_BIN")); isExec(override) {
		return override, nil
	}
	if local := localBuildFallback(); local != "" {
		return local, nil
	}
	if err := downloadEngine(version, installed); err != nil {
		return "", err
	}
	if isExec(installed) {
		return installed, nil
	}
	return "", fmt.Errorf(
		"could not resolve the zrt-console engine v%s: set ZRT_CONSOLE_BIN to a built binary, "+
			"or ensure network access to the engine CDN", version)
}
