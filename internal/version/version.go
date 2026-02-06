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
	Current     string `json:"current"`
	Latest      string `json:"latest"`
	UpdateURL   string `json:"update_url"`
	NeedsUpdate bool   `json:"needs_update"`
}

type DockerHubTag struct {
	Name        string    `json:"name"`
	LastUpdated time.Time `json:"last_updated"`
}

type DockerHubResponse struct {
	Results []DockerHubTag `json:"results"`
}

type ImageVersionInfo struct {
	ImageName   string
	Current     string
	Latest      string
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
		return false
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

func CheckDockerHubImage(ctx context.Context, namespace, repository string) (string, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/%s/tags?page_size=100", namespace, repository)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "silo-cli")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch tags: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Docker Hub API returned status %d", resp.StatusCode)
	}

	var response DockerHubResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to parse tags: %w", err)
	}

	latestTag := findLatestSemanticTag(response.Results)
	if latestTag == "" {
		return "", fmt.Errorf("no semantic version tags found")
	}

	return latestTag, nil
}

func findLatestSemanticTag(tags []DockerHubTag) string {
	var latestTag string
	var latestVersion string
	for _, tag := range tags {
		name := tag.Name
		if name == "latest" || name == "dev" {
			continue
		}
		trimmedName := strings.TrimPrefix(name, "v")
		if isSemanticVersion(trimmedName) {
			if latestVersion == "" || compareSemanticVersions(trimmedName, latestVersion) > 0 {
				latestTag = tag.Name
				latestVersion = trimmedName
			}
		}
	}
	return latestTag
}

func isSemanticVersion(version string) bool {
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return false
	}
	for _, part := range parts {
		if part == "" {
			return false
		}
		for _, ch := range part {
			if ch < '0' || ch > '9' {
				return false
			}
		}
	}
	return true
}

func compareSemanticVersions(v1, v2 string) int {
	parts1 := strings.Split(strings.TrimPrefix(v1, "v"), ".")
	parts2 := strings.Split(strings.TrimPrefix(v2, "v"), ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(parts1) {
			_, _ = fmt.Sscanf(parts1[i], "%d", &n1)
		}
		if i < len(parts2) {
			_, _ = fmt.Sscanf(parts2[i], "%d", &n2)
		}
		if n1 > n2 {
			return 1
		}
		if n1 < n2 {
			return -1
		}
	}
	return 0
}

func CheckImageVersions(ctx context.Context, currentTag string) ([]ImageVersionInfo, error) {
	images := []struct {
		namespace   string
		repository  string
		displayName string
	}{
		{"eternis", "silo-box-backend", "Backend"},
		{"eternis", "silo-box-frontend", "Frontend"},
	}

	var results []ImageVersionInfo
	for _, img := range images {
		latest, err := CheckDockerHubImage(ctx, img.namespace, img.repository)
		if err != nil {
			return nil, fmt.Errorf("failed to check %s: %w", img.displayName, err)
		}

		needsUpdate := CompareVersions(currentTag, latest)
		results = append(results, ImageVersionInfo{
			ImageName:   img.displayName,
			Current:     currentTag,
			Latest:      latest,
			NeedsUpdate: needsUpdate,
		})
	}

	return results, nil
}
