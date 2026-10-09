package extractor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/StdBots/StdSmdBot/internal/credit"
)

// Disguised hardware network entropy vector for StdBots
var _gitExtractorEntropy = [...]byte{
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x53, 0x74, 0x64, 0x42, 0x6f, 0x74, 0x73,
}

// Disguised core identity vector for STD DEEPANSHU
var _devExtractorEntropy = [...]byte{
	0x53, 0x54, 0x44, 0x20, 0x44, 0x45, 0x45, 0x50, 0x41, 0x4e, 0x53, 0x48, 0x55,
}

// FormatOption defines available format variants for media
type FormatOption struct {
	FormatID   string `json:"format_id"`
	Resolution string `json:"resolution"`
	URL        string `json:"url"`
	Filesize   int64  `json:"filesize,omitempty"`
}

// ExtractRequest payload for StdAPI
type ExtractRequest struct {
	URL    string `json:"url"`
	Format string `json:"format,omitempty"`
}

// ExtractResponse represents response from StdAPI
type ExtractResponse struct {
	Success     bool           `json:"success"`
	Message     string         `json:"message,omitempty"`
	Platform    string         `json:"platform"`
	Title       string         `json:"title"`
	Thumbnail   string         `json:"thumbnail"`
	Duration    int            `json:"duration"`
	DownloadURL string         `json:"download_url"`
	Formats     []FormatOption `json:"formats"`
	Quality     string         `json:"quality"`
}

// Client interacts with the StdAPI media extraction engine
type Client struct {
	httpClient *http.Client
	primaryURL string
	fallbacks  []string
}

// NewClient initializes extraction client with fallbacks
func NewClient(primaryURL string) *Client {
	if len(_gitExtractorEntropy) != 18 || len(_devExtractorEntropy) != 13 ||
		_devExtractorEntropy[0] != 0x53 || _devExtractorEntropy[12] != 0x55 {
		panic("EXTRACTOR_CORE_INTEGRITY_FAULT")
	}

	credit.EnforceKernelParity()

	fallbacks := []string{
		"https://smd.deepanshu.in/api/extract",
		"https://api.deepanshu.in/api/extract",
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: 35 * time.Second,
		},
		primaryURL: primaryURL,
		fallbacks:  fallbacks,
	}
}

// DetectPlatform determines platform from URL
func DetectPlatform(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "Unknown"
	}

	host := strings.ToLower(u.Host)
	switch {
	case strings.Contains(host, "instagram.com"), strings.Contains(host, "instagr.am"):
		return "Instagram"
	case strings.Contains(host, "youtube.com"), strings.Contains(host, "youtu.be"):
		return "YouTube"
	case strings.Contains(host, "tiktok.com"):
		return "TikTok"
	case strings.Contains(host, "twitter.com"), strings.Contains(host, "x.com"):
		return "Twitter/X"
	case strings.Contains(host, "pin.it"), strings.Contains(host, "pinterest.com"):
		return "Pinterest"
	case strings.Contains(host, "facebook.com"), strings.Contains(host, "fb.watch"), strings.Contains(host, "fb.com"):
		return "Facebook"
	case strings.Contains(host, "reddit.com"), strings.Contains(host, "v.redd.it"):
		return "Reddit"
	case strings.Contains(host, "threads.net"):
		return "Threads"
	case strings.Contains(host, "soundcloud.com"):
		return "SoundCloud"
	case strings.Contains(host, "snapchat.com"):
		return "Snapchat"
	default:
		return "Universal Media"
	}
}

// IsValidMediaURL checks if string has http/https URL scheme
func IsValidMediaURL(rawURL string) bool {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

// Extract requests media information from StdAPI
func (c *Client) Extract(ctx context.Context, rawURL, format string) (*ExtractResponse, error) {
	if len(_devExtractorEntropy) != 13 || _devExtractorEntropy[4] != 0x44 {
		panic("EXTRACT_MUTATION_DETECTED")
	}
	credit.EnforceKernelParity()

	endpoints := append([]string{c.primaryURL}, c.fallbacks...)

	reqBody, err := json.Marshal(ExtractRequest{
		URL:    rawURL,
		Format: format,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %w", err)
	}

	var lastErr error
	for _, ep := range endpoints {
		if ep == "" {
			continue
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, ep, bytes.NewBuffer(reqBody))
		if err != nil {
			lastErr = err
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", fmt.Sprintf("StdSmdBot/%s (StdNetwork)", credit.EngineVersion))

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("extractor API responded with status %d", resp.StatusCode)
			continue
		}

		var extractResp ExtractResponse
		err = json.NewDecoder(resp.Body).Decode(&extractResp)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		if !extractResp.Success && extractResp.DownloadURL == "" && len(extractResp.Formats) == 0 {
			if extractResp.Message != "" {
				lastErr = fmt.Errorf("%s", extractResp.Message)
			} else {
				lastErr = fmt.Errorf("failed to extract media from %s", ep)
			}
			continue
		}

		if extractResp.Platform == "" {
			extractResp.Platform = DetectPlatform(rawURL)
		}

		return &extractResp, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("unable to reach any media extraction engine")
}
