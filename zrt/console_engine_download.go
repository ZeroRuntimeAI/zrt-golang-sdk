package zrt

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const defaultEngineCDN = "https://cdn.videosdk.live/agent/zrt-console"

// engineCDNBase returns the CDN base URL, overridable via ZRT_ENGINE_CDN.
func engineCDNBase() string {
	base := strings.TrimSpace(os.Getenv("ZRT_ENGINE_CDN"))
	if base == "" {
		base = defaultEngineCDN
	}
	return strings.TrimRight(base, "/")
}

// engineAssetURL returns the download URL for the engine tarball.
func engineAssetURL(version, target string) string {
	return fmt.Sprintf("%s/%s/zrt-console-%s.tar.gz", engineCDNBase(), version, target)
}

// fetchToFile downloads url into dest.
func fetchToFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed (%d) for %s", resp.StatusCode, url)
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

// verifySHA256 checks that tarPath hashes to the value in shaPath (first token).
func verifySHA256(tarPath, shaPath string) error {
	raw, err := os.ReadFile(shaPath)
	if err != nil {
		return err
	}
	fields := strings.Fields(string(raw))
	if len(fields) == 0 {
		return fmt.Errorf("empty checksum file")
	}
	expected := strings.ToLower(fields[0])
	f, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	actual := hex.EncodeToString(h.Sum(nil))
	if expected != actual {
		return fmt.Errorf("engine checksum mismatch: expected %s, got %s", expected, actual)
	}
	return nil
}

// extractEngine pulls the engine binary out of the gzipped tar at tarPath and
// writes it to destBinaryPath.
func extractEngine(tarPath, destBinaryPath string) error {
	name := engineBinaryName()
	f, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if filepath.Base(hdr.Name) != name {
			continue
		}
		out, err := os.Create(destBinaryPath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return err
		}
		out.Close()
		return nil
	}
	return fmt.Errorf("archive did not contain %q", name)
}

// downloadEngine fetches, verifies, and installs the engine binary at
// destBinaryPath. Checksum verification is best-effort except on mismatch.
func downloadEngine(version, destBinaryPath string) error {
	target, err := detectTarget()
	if err != nil {
		return err
	}
	assetURL := engineAssetURL(version, target)
	if err := os.MkdirAll(filepath.Dir(destBinaryPath), 0o755); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp("", "zrt-engine-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	tarPath := filepath.Join(tmp, "engine.tar.gz")
	if err := fetchToFile(assetURL, tarPath); err != nil {
		return err
	}

	shaPath := filepath.Join(tmp, "engine.tar.gz.sha256")
	if shaErr := fetchToFile(assetURL+".sha256", shaPath); shaErr == nil {
		// A verified mismatch is fatal; failure to obtain the checksum is not.
		if err := verifySHA256(tarPath, shaPath); err != nil {
			return err
		}
	}

	if err := extractEngine(tarPath, destBinaryPath); err != nil {
		return err
	}
	return os.Chmod(destBinaryPath, 0o755)
}
