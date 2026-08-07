package updater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Result struct {
	Current   string
	Latest    string
	URL       string
	Notes     string
	Available bool
}

type Checker struct {
	Client *http.Client
}

func (c Checker) Check(ctx context.Context, current, updateURL, repository string) (Result, error) {
	endpoint := strings.TrimSpace(updateURL)
	if endpoint == "" && strings.TrimSpace(repository) != "" {
		endpoint = "https://api.github.com/repos/" + strings.Trim(strings.TrimSpace(repository), "/") + "/releases/latest"
	}
	if endpoint == "" {
		return Result{Current: current}, errors.New("update source is not configured; set SCROBBLER_UPDATE_URL or build with a repository")
	}
	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Result{Current: current}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json, application/json, text/plain")
	req.Header.Set("User-Agent", "lastfm-scrobbler-update-checker")
	resp, err := client.Do(req)
	if err != nil {
		return Result{Current: current}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return Result{Current: current}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Result{Current: current}, fmt.Errorf("update server returned HTTP %d", resp.StatusCode)
	}
	latest, pageURL, notes := parseResponse(data)
	if latest == "" {
		return Result{Current: current}, errors.New("update response did not contain a version")
	}
	return Result{
		Current:   current,
		Latest:    latest,
		URL:       pageURL,
		Notes:     notes,
		Available: compareVersions(latest, current) > 0,
	}, nil
}

func parseResponse(data []byte) (latest, pageURL, notes string) {
	var payload struct {
		Version string `json:"version"`
		TagName string `json:"tag_name"`
		URL     string `json:"url"`
		HTMLURL string `json:"html_url"`
		Notes   string `json:"notes"`
		Body    string `json:"body"`
	}
	if json.Unmarshal(data, &payload) == nil {
		latest = first(payload.Version, payload.TagName)
		pageURL = first(payload.HTMLURL, payload.URL)
		notes = first(payload.Notes, payload.Body)
		if latest != "" {
			return strings.TrimSpace(latest), strings.TrimSpace(pageURL), strings.TrimSpace(notes)
		}
	}
	return strings.TrimSpace(string(data)), "", ""
}

func first(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

var numericPart = regexp.MustCompile(`\d+`)

func compareVersions(a, b string) int {
	parts := func(value string) []int {
		matches := numericPart.FindAllString(strings.TrimPrefix(strings.TrimSpace(value), "v"), -1)
		out := make([]int, 0, len(matches))
		for _, match := range matches {
			number, _ := strconv.Atoi(match)
			out = append(out, number)
		}
		return out
	}
	left, right := parts(a), parts(b)
	length := len(left)
	if len(right) > length {
		length = len(right)
	}
	for i := 0; i < length; i++ {
		var l, r int
		if i < len(left) {
			l = left[i]
		}
		if i < len(right) {
			r = right[i]
		}
		if l > r {
			return 1
		}
		if l < r {
			return -1
		}
	}
	return 0
}
