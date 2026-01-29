package version

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type ReleaseInfo struct {
	TagName     string    `json:"tag_name"`
	HTMLURL     string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
}

type VersionInfo struct {
	Current     string
	Latest      string
	UpdateURL   string
	NeedsUpdate bool
}

func CheckLatestRelease(ctx context.Context) (*ReleaseInfo, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/repos/EternisAI/silo/releases/latest", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "silo-cli")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	return &release, nil
}

func CompareVersions(current, latest string) bool {
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")

	if current == "dev" {
		return true
	}

	return current != latest
}

func Check(ctx context.Context, currentVersion string) (*VersionInfo, error) {
	release, err := CheckLatestRelease(ctx)
	if err != nil {
		return nil, err
	}

	needsUpdate := CompareVersions(currentVersion, release.TagName)

	return &VersionInfo{
		Current:     currentVersion,
		Latest:      release.TagName,
		UpdateURL:   release.HTMLURL,
		NeedsUpdate: needsUpdate,
	}, nil
}
