package connection

import (
	"context"
	"strings"
	"time"

	"github.com/deathrashed/lastfm-scrobbler/internal/config"
	"github.com/deathrashed/lastfm-scrobbler/internal/lastfm"
)

type Item struct {
	Label   string `json:"label"`
	OK      bool   `json:"ok"`
	Detail  string `json:"detail"`
	Skipped bool   `json:"skipped,omitempty"`
}

type Report struct {
	Items []Item `json:"items"`
}

func (r Report) OK() bool {
	for _, item := range r.Items {
		if !item.OK && !item.Skipped {
			return false
		}
	}
	return len(r.Items) > 0
}

func Test(ctx context.Context, cfg config.Config, client lastfm.Client) Report {
	report := Report{}
	apiKey := strings.TrimSpace(cfg.APIKey)
	report.Items = append(report.Items, Item{Label: "API KEY", OK: apiKey != "", Detail: configured(apiKey)})
	if apiKey == "" {
		report.Items = append(report.Items,
			Item{Label: "READ API", Detail: "not tested", Skipped: true},
			Item{Label: "AUTH", Detail: "not tested", Skipped: true},
		)
		return report
	}

	readCtx, cancelRead := context.WithTimeout(ctx, 20*time.Second)
	_, readErr := client.SearchAlbums(readCtx, "Metallica")
	cancelRead()
	if readErr != nil {
		report.Items = append(report.Items, Item{Label: "READ API", Detail: readErr.Error()})
	} else {
		report.Items = append(report.Items, Item{Label: "READ API", OK: true, Detail: "Last.fm lookup succeeded"})
	}

	secretOK := strings.TrimSpace(cfg.APISecret) != ""
	report.Items = append(report.Items, Item{Label: "API SECRET", OK: secretOK, Detail: configured(cfg.APISecret)})
	if !secretOK {
		report.Items = append(report.Items, Item{Label: "AUTH", Detail: "API secret is required", Skipped: true})
		return report
	}

	hasSession := strings.TrimSpace(cfg.SessionKey) != ""
	hasPassword := strings.TrimSpace(cfg.Username) != "" && strings.TrimSpace(cfg.Password) != ""
	if hasSession {
		// Last.fm has no side-effect-free write call. The configured session
		// key will be validated by the first signed operation, so do not claim
		// that this read-only connection test has proven it is still active.
		report.Items = append(report.Items, Item{Label: "AUTH", OK: true, Detail: "session key configured; validated on first signed write"})
		return report
	}
	if !hasPassword {
		report.Items = append(report.Items, Item{Label: "AUTH", Detail: "configure a session key or username/password"})
		return report
	}

	authCtx, cancelAuth := context.WithTimeout(ctx, 30*time.Second)
	authErr := client.Authenticate(authCtx)
	cancelAuth()
	if authErr != nil {
		report.Items = append(report.Items, Item{Label: "AUTH", Detail: authErr.Error()})
	} else {
		if sessionClient, ok := client.(interface{ SessionKey() string }); ok {
			_ = config.PersistSessionKey(cfg, sessionClient.SessionKey())
		}
		report.Items = append(report.Items, Item{Label: "AUTH", OK: true, Detail: "mobile session obtained with username + password"})
	}
	return report
}

func configured(value string) string {
	if strings.TrimSpace(value) == "" {
		return "missing"
	}
	return "configured"
}
