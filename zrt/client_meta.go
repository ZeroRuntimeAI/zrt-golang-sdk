package zrt

import (
	"os"
	"runtime"
	"strings"
	"sync"

	pb "github.com/ZeroRuntimeAI/zrt-golang-sdk/internal/pb"
	"google.golang.org/grpc/metadata"
)

const (
	// protoVersion is sent in the x-zrt-proto-version header and VersionInfo.ProtoVersion.
	protoVersion = "1"
	// minimumRuntimeVersion is the lowest runtime version this SDK supports.
	minimumRuntimeVersion = "0.0.1"
	sdkName               = "zrt-golang-sdk"
)

func sdkVersion() string {
	if Version != "" {
		return Version
	}
	return "0.0.0"
}

// clientVersionInfo returns the version info sent in session config and registration.
func clientVersionInfo() *pb.VersionInfo {
	return &pb.VersionInfo{
		SdkVersion:            sdkVersion(),
		MinimumRuntimeVersion: minimumRuntimeVersion,
		ProtoVersion:          protoVersion,
	}
}

func deviceOS() string {
	return strings.ToLower(strings.TrimSpace(runtime.GOOS))
}

func deviceArch() string {
	return strings.ToLower(runtime.GOARCH)
}

func hostName() string {
	h, err := os.Hostname()
	if err != nil {
		return ""
	}
	return h
}

var (
	staticHeadersOnce sync.Once
	staticHeaders     []string // flat key,value pairs
)

func staticHeaderPairs() []string {
	staticHeadersOnce.Do(func() {
		osHeader := strings.ToLower(strings.TrimSpace(deviceOS() + " " + osRelease()))
		staticHeaders = []string{
			"x-zrt-sdk", sdkName,
			"x-zrt-sdk-version", sdkVersion(),
			"x-zrt-proto-version", protoVersion,
			"x-zrt-device-os", osHeader,
			"x-zrt-device-arch", deviceArch(),
			"x-zrt-host", hostName(),
		}
	})
	return staticHeaders
}

// osRelease reports an OS release string for the device-os header. No portable
// source exists, so the Go runtime version is used (header is informational).
func osRelease() string {
	return strings.TrimPrefix(runtime.Version(), "go")
}

// clientMetadata returns the gRPC metadata for an outgoing call.
func clientMetadata(authToken string) metadata.MD {
	pairs := staticHeaderPairs()
	md := metadata.Pairs(pairs...)
	token := authToken
	if token == "" {
		token = os.Getenv("ZRT_AUTH_TOKEN")
	}
	if token != "" {
		md.Set("authorization", "Bearer "+token)
	}
	return md
}
