package tui

import (
	"context"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/deathrashed/lastfm-scrobbler/internal/lastfm"
	"github.com/deathrashed/lastfm-scrobbler/internal/theme"
)

type activityState uint8

const (
	activityLoading activityState = iota
	activityCurrent
	activityRecent
	activityNoTracks
	activityUnavailable
)

var activityVolumeFrames = []string{"", "", "", "", "", ""}

type activityResultMsg struct {
	seq    uint64
	tracks []lastfm.RecentTrack
	err    error
}

type activityRefreshMsg struct{ seq uint64 }
type activityAnimationMsg struct{ seq uint64 }

func (m model) activityPollingEnabled() bool {
	return m.nowPlayingEnabled() && strings.TrimSpace(m.cfg.Username) != "" && m.client != nil
}

func (m model) activityFetchCmd() tea.Cmd {
	if !m.activityPollingEnabled() {
		return nil
	}
	seq := m.activitySeq
	client := m.client
	username := m.cfg.Username
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		tracks, err := client.GetRecentTracks(ctx, username, time.Time{})
		return activityResultMsg{seq: seq, tracks: tracks, err: err}
	}
}

func activityRefreshCmd(seq uint64) tea.Cmd {
	return tea.Tick(30*time.Second, func(time.Time) tea.Msg { return activityRefreshMsg{seq: seq} })
}

func activityAnimationCmd(seq uint64) tea.Cmd {
	return tea.Tick(280*time.Millisecond, func(time.Time) tea.Msg { return activityAnimationMsg{seq: seq} })
}

func (m model) activityContent() string {
	if !m.nowPlayingEnabled() {
		return ""
	}
	if m.activityState == activityLoading && !m.activityPollingEnabled() {
		return theme.MutedStyle.Render("Last.fm activity unavailable")
	}
	switch m.activityState {
	case activityLoading:
		return theme.MutedStyle.Render("loading Last.fm activity")
	case activityNoTracks:
		return theme.MutedStyle.Render("no recent scrobbles")
	case activityUnavailable:
		return theme.MutedStyle.Render("Last.fm activity unavailable")
	case activityCurrent, activityRecent:
		artist, title := activityTextParts(m.activityTrack.Artist, m.activityTrack.Title, maxInt(1, m.appWidth()-2-displayWidth(activityVolumeFrames[0])-2))
		icon := ""
		iconStyle := theme.IconStyle
		if m.activityState == activityCurrent {
			icon = activityVolumeFrames[m.activityFrame%len(activityVolumeFrames)]
			iconStyle = theme.AccentTextStyle
		}
		return iconStyle.Render(icon) + "  " + theme.AlbumStyle.Render(artist) +
			theme.MutedStyle.Render(" - ") + theme.PrimaryTextStyle.Render(title)
	default:
		return ""
	}
}

func (m model) activityShouldAnimate() bool {
	return m.activityPollingEnabled() && m.activityState == activityCurrent
}

func activityTextParts(artist, title string, width int) (string, string) {
	artist = strings.TrimSpace(artist)
	title = strings.TrimSpace(title)
	const separatorWidth = 3
	width = maxInt(separatorWidth+2, width)
	if displayWidth(artist)+separatorWidth+displayWidth(title) <= width {
		return artist, title
	}
	partWidth := maxInt(1, (width-separatorWidth)/2)
	artist = truncateToWidth(artist, partWidth)
	title = truncateToWidth(title, maxInt(1, width-separatorWidth-displayWidth(artist)))
	return artist, title
}
