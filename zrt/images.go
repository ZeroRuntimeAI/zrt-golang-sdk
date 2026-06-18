package zrt

// Image-source helpers for building an ImageContent from local data. The runtime
// forwards the resulting URL (an http(s) URL or an embedded base64 "data:" URL) to the
// LLM as a vision input (Gemini decodes it to inline data; OpenAI/Anthropic accept it
// directly). Mirrors the JS SDK's ImageContent.from* helpers (cross-SDK parity).

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"os"
)

// ImageContentFromURL builds an ImageContent from an http(s)/gs URL the LLM fetches.
func ImageContentFromURL(url string, detail ...string) ImageContent {
	return ImageContent{Type: "image", URL: url, Detail: detailOr(detail)}
}

// ImageContentFromDataURL builds an ImageContent from an existing "data:" URL
// (passed through unchanged).
func ImageContentFromDataURL(dataURL string, detail ...string) ImageContent {
	return ImageContent{Type: "image", URL: dataURL, Detail: detailOr(detail)}
}

// ImageContentFromBase64 builds an ImageContent from a base64 image string and its
// MIME type (e.g. "image/png").
func ImageContentFromBase64(b64, mimeType string, detail ...string) ImageContent {
	return ImageContent{Type: "image", URL: "data:" + mimeType + ";base64," + b64, Detail: detailOr(detail)}
}

// ImageContentFromBytes builds an ImageContent from already-encoded image bytes
// (e.g. a captured frame's data) and its MIME type.
func ImageContentFromBytes(data []byte, mimeType string, detail ...string) ImageContent {
	return ImageContentFromBase64(base64.StdEncoding.EncodeToString(data), mimeType, detail...)
}

// ImageContentFromFile reads an image file and embeds it as a "data:" URL, sniffing the
// MIME type from the file's content.
func ImageContentFromFile(path string, detail ...string) (ImageContent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ImageContent{}, err
	}
	mime := detectImageMIME(data)
	if mime == "" {
		return ImageContent{}, fmt.Errorf("could not determine image type of %s (expected JPEG/PNG/GIF/WebP/BMP)", path)
	}
	return ImageContentFromBytes(data, mime, detail...), nil
}

// ImageContentFromImage encodes a decoded image (the PIL / AV-frame analog — also how
// raw pixels become an image via image.RGBA) to PNG and embeds it as a "data:" URL.
func ImageContentFromImage(img image.Image, detail ...string) (ImageContent, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return ImageContent{}, err
	}
	return ImageContentFromBytes(buf.Bytes(), "image/png", detail...), nil
}

func detailOr(detail []string) string {
	if len(detail) > 0 && detail[0] != "" {
		return detail[0]
	}
	return "auto"
}

// detectImageMIME sniffs the MIME type of encoded image bytes from their magic-number
// signature. Recognizes JPEG, PNG, GIF, WebP, and BMP; returns "" if unrecognized.
func detectImageMIME(data []byte) string {
	if len(data) < 4 {
		return ""
	}
	switch {
	case data[0] == 0xff && data[1] == 0xd8 && data[2] == 0xff:
		return "image/jpeg"
	case data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4e && data[3] == 0x47:
		return "image/png"
	case data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x38:
		return "image/gif"
	case data[0] == 0x42 && data[1] == 0x4d:
		return "image/bmp"
	case len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP":
		return "image/webp"
	}
	return ""
}
