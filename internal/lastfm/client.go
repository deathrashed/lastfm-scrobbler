package lastfm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var ErrAlbumNotFound = errors.New("album not found")

type ErrWithSuggestion struct {
	Err        error
	Suggestion string
}

func (e *ErrWithSuggestion) Error() string {
	if e.Suggestion != "" {
		return fmt.Sprintf("%v (did you mean %q?)", e.Err, e.Suggestion)
	}
	return e.Err.Error()
}
func (e *ErrWithSuggestion) Unwrap() error { return e.Err }

type Client interface {
	Authenticate(ctx context.Context) error
	SearchAlbums(ctx context.Context, query string) ([]Album, error)
	GetAlbumTracks(ctx context.Context, artist, album string) (Album, error)
	GetDiscography(ctx context.Context, artist string) ([]Album, error)
	GetSimilarAlbums(ctx context.Context, artist string, limit int) ([]Album, error)
	GetRecentTracks(ctx context.Context, user string, from time.Time) ([]RecentTrack, error)
	Scrobble(ctx context.Context, artist, track, album string, ts int64) error
}

const apiBase = "https://ws.audioscrobbler.com/2.0/"

func New(apiKey, apiSecret, username, password, sessionKey string) Client {
	return &client{
		apiKey:     strings.TrimSpace(apiKey),
		apiSecret:  strings.TrimSpace(apiSecret),
		username:   strings.TrimSpace(username),
		password:   password,
		sessionKey: strings.TrimSpace(sessionKey),
		baseURL:    apiBase,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type client struct {
	apiKey     string
	apiSecret  string
	username   string
	password   string
	sessionKey string
	baseURL    string
	httpClient *http.Client
}

func (c *client) SessionKey() string { return c.sessionKey }

func (c *client) doRequest(ctx context.Context, input map[string]string, v interface{}) error {
	if c.apiKey == "" {
		return errors.New("last.fm API key is not configured; set API_KEY or LASTFM_API_KEY")
	}

	params := make(map[string]string, len(input)+3)
	for key, value := range input {
		params[key] = value
	}

	methodName := params["method"]
	signed := methodName == "auth.getMobileSession" || methodName == "track.scrobble"
	httpMethod := http.MethodGet
	params["api_key"] = c.apiKey
	if signed {
		if c.apiSecret == "" {
			return errors.New("last.fm API secret is not configured; set API_SECRET or LASTFM_API_SECRET")
		}
		params["api_sig"] = sign(params, c.apiSecret)
		httpMethod = http.MethodPost
	}
	params["format"] = "json"

	baseURL := c.baseURL
	if baseURL == "" {
		baseURL = apiBase
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return err
	}

	var body io.Reader
	if httpMethod == http.MethodPost {
		values := url.Values{}
		for key, value := range params {
			values.Set(key, value)
		}
		body = strings.NewReader(values.Encode())
	} else {
		query := u.Query()
		for key, value := range params {
			query.Set(key, value)
		}
		u.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, httpMethod, u.String(), body)
	if err != nil {
		return err
	}
	if httpMethod == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiError struct {
		Error   int    `json:"error"`
		Message string `json:"message"`
	}
	if json.Unmarshal(responseBody, &apiError) == nil && apiError.Error != 0 {
		return fmt.Errorf("last.fm API error %d: %s", apiError.Error, apiError.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("last.fm API returned HTTP %d: %s", resp.StatusCode, compactBody(responseBody))
	}
	if err := json.Unmarshal(responseBody, v); err != nil {
		return fmt.Errorf("decode Last.fm response: %w", err)
	}
	return nil
}

func compactBody(body []byte) string {
	body = bytes.TrimSpace(body)
	if len(body) > 240 {
		body = body[:240]
	}
	return string(body)
}

func (c *client) Authenticate(ctx context.Context) error {
	// A persisted web-service session key is sufficient for authenticated
	// scrobbling and avoids storing the account password.
	if strings.TrimSpace(c.sessionKey) != "" {
		return nil
	}
	if c.username == "" {
		return errors.New("last.fm username is not configured; set LASTFM_USERNAME")
	}
	if strings.TrimSpace(c.password) == "" {
		return errors.New("last.fm password is not configured; set LASTFM_PASSWORD")
	}

	params := map[string]string{
		"method":   "auth.getMobileSession",
		"username": c.username,
		// The current Last.fm mobile-session API requires plain text here.
		"password": c.password,
	}
	var response struct {
		Session struct {
			Key string `json:"key"`
		} `json:"session"`
	}
	if err := c.doRequest(ctx, params, &response); err != nil {
		return err
	}
	if response.Session.Key == "" {
		return errors.New("last.fm returned an empty session key")
	}
	c.sessionKey = response.Session.Key
	return nil
}

func (c *client) SearchAlbums(ctx context.Context, query string) ([]Album, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, errors.New("album search query is empty")
	}
	params := map[string]string{
		"method": "album.search",
		"album":  query,
		"limit":  "50",
	}
	var response struct {
		Results struct {
			AlbumMatches struct {
				Albums []struct {
					Name   string    `json:"name"`
					Title  string    `json:"title"`
					Artist textValue `json:"artist"`
				} `json:"album"`
			} `json:"albummatches"`
		} `json:"results"`
	}
	if err := c.doRequest(ctx, params, &response); err != nil {
		return nil, err
	}

	albums := make([]Album, 0, len(response.Results.AlbumMatches.Albums))
	seen := map[string]bool{}
	for _, item := range response.Results.AlbumMatches.Albums {
		title := firstNonEmpty(item.Name, item.Title)
		artist := item.Artist.String()
		if title == "" || artist == "" {
			continue
		}
		key := strings.ToLower(artist + "\x00" + title)
		if seen[key] {
			continue
		}
		seen[key] = true
		albums = append(albums, Album{Artist: artist, Title: title})
	}
	return albums, nil
}

func (c *client) GetAlbumTracks(ctx context.Context, artist, albumName string) (Album, error) {
	result, err := c.fetchAlbumTracks(ctx, artist, albumName)
	if err == nil {
		return result, nil
	}
	if !isAlbumNotFound(err) {
		return Album{}, err
	}

	candidates, searchErr := c.SearchAlbums(ctx, albumName)
	if searchErr != nil || len(candidates) == 0 {
		return Album{}, ErrAlbumNotFound
	}
	best := bestFuzzyMatch(albumName, candidates)
	if best == "" {
		return Album{}, ErrAlbumNotFound
	}
	result, suggestErr := c.fetchAlbumTracks(ctx, artist, best)
	if suggestErr != nil {
		return Album{}, ErrAlbumNotFound
	}
	// A usable tracklist is more important to the TUI than interrupting the
	// workflow with a suggestion error.
	return result, nil
}

func (c *client) fetchAlbumTracks(ctx context.Context, artist, albumName string) (Album, error) {
	artist = strings.TrimSpace(artist)
	albumName = strings.TrimSpace(albumName)
	if artist == "" {
		return Album{}, errors.New("artist name is empty")
	}
	if albumName == "" {
		return Album{}, errors.New("album name is empty")
	}

	params := map[string]string{
		"method":      "album.getInfo",
		"artist":      artist,
		"album":       albumName,
		"autocorrect": "1",
	}
	var response struct {
		Album struct {
			Name   string    `json:"name"`
			Artist textValue `json:"artist"`
			Tracks struct {
				Track trackValues `json:"track"`
			} `json:"tracks"`
		} `json:"album"`
	}
	if err := c.doRequest(ctx, params, &response); err != nil {
		return Album{}, err
	}
	if strings.TrimSpace(response.Album.Name) == "" {
		return Album{}, ErrAlbumNotFound
	}

	album := Album{
		Artist: firstNonEmpty(response.Album.Artist.String(), artist),
		Title:  strings.TrimSpace(response.Album.Name),
	}
	for _, item := range response.Album.Tracks.Track {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			continue
		}
		album.Tracks = append(album.Tracks, Track{
			Title:    name,
			Duration: int(item.Duration), // Last.fm reports seconds.
		})
	}
	if len(album.Tracks) == 0 {
		return Album{}, fmt.Errorf("%w: no tracks returned for %s — %s", ErrAlbumNotFound, artist, albumName)
	}
	return album, nil
}

func isAlbumNotFound(err error) bool {
	if errors.Is(err, ErrAlbumNotFound) {
		return true
	}
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "album not found") || strings.Contains(message, "invalid resource")
}

func bestFuzzyMatch(query string, candidates []Album) string {
	clean := func(value string) []string {
		replacer := strings.NewReplacer("'", "", ",", "", ";", "", "-", " ", "—", " ")
		return strings.Fields(strings.ToLower(replacer.Replace(value)))
	}
	queryWords := clean(query)
	if len(queryWords) == 0 {
		return ""
	}

	bestScore := 0.0
	bestTitle := ""
	for _, candidate := range candidates {
		candidateWords := clean(candidate.Title)
		if len(candidateWords) == 0 {
			continue
		}
		set := map[string]bool{}
		for _, word := range candidateWords {
			set[word] = true
		}
		common := 0
		for _, word := range queryWords {
			if set[word] {
				common++
			}
		}
		score := float64(common) / float64(max(len(queryWords), len(candidateWords)))
		if score > bestScore && score > 0.5 {
			bestScore = score
			bestTitle = candidate.Title
		}
	}
	return bestTitle
}

func (c *client) GetDiscography(ctx context.Context, artist string) ([]Album, error) {
	artist = strings.TrimSpace(artist)
	if artist == "" {
		return nil, errors.New("artist name is empty")
	}
	params := map[string]string{
		"method":      "artist.getTopAlbums",
		"artist":      artist,
		"autocorrect": "1",
		"limit":       "200",
	}
	var response struct {
		TopAlbums struct {
			Albums []struct {
				Name   string    `json:"name"`
				Title  string    `json:"title"`
				Artist textValue `json:"artist"`
			} `json:"album"`
		} `json:"topalbums"`
	}
	if err := c.doRequest(ctx, params, &response); err != nil {
		return nil, err
	}

	albums := make([]Album, 0, len(response.TopAlbums.Albums))
	seen := map[string]bool{}
	for _, item := range response.TopAlbums.Albums {
		title := firstNonEmpty(item.Name, item.Title)
		albumArtist := firstNonEmpty(item.Artist.String(), artist)
		if title == "" {
			continue
		}
		key := strings.ToLower(albumArtist + "\x00" + title)
		if seen[key] {
			continue
		}
		seen[key] = true
		albums = append(albums, Album{Artist: albumArtist, Title: title})
	}
	if len(albums) == 0 {
		return nil, fmt.Errorf("no albums found for %s", artist)
	}
	return albums, nil
}

func (c *client) getTopAlbums(ctx context.Context, artist string, limit int) ([]Album, error) {
	artist = strings.TrimSpace(artist)
	if artist == "" {
		return nil, errors.New("artist name is empty")
	}
	if limit < 1 {
		limit = 1
	}
	params := map[string]string{
		"method":      "artist.getTopAlbums",
		"artist":      artist,
		"autocorrect": "1",
		"limit":       strconv.Itoa(limit),
	}
	var response struct {
		TopAlbums struct {
			Albums []struct {
				Name   string    `json:"name"`
				Title  string    `json:"title"`
				Artist textValue `json:"artist"`
			} `json:"album"`
		} `json:"topalbums"`
	}
	if err := c.doRequest(ctx, params, &response); err != nil {
		return nil, err
	}
	var albums []Album
	for _, item := range response.TopAlbums.Albums {
		title := firstNonEmpty(item.Name, item.Title)
		if title == "" {
			continue
		}
		albums = append(albums, Album{Artist: firstNonEmpty(item.Artist.String(), artist), Title: title})
	}
	return albums, nil
}

// GetSimilarAlbums derives album recommendations from artists returned by
// artist.getSimilar, then takes the top album for each similar artist.
func (c *client) GetSimilarAlbums(ctx context.Context, artist string, limit int) ([]Album, error) {
	artist = strings.TrimSpace(artist)
	if artist == "" {
		return nil, errors.New("artist name is empty")
	}
	if limit < 1 {
		limit = 12
	}
	if limit > 25 {
		limit = 25
	}
	params := map[string]string{
		"method":      "artist.getSimilar",
		"artist":      artist,
		"autocorrect": "1",
		"limit":       strconv.Itoa(limit),
	}
	var response struct {
		Similar struct {
			Artists []struct {
				Name  string      `json:"name"`
				Match floatString `json:"match"`
			} `json:"artist"`
		} `json:"similarartists"`
	}
	if err := c.doRequest(ctx, params, &response); err != nil {
		return nil, err
	}

	seen := map[string]bool{}
	albums := make([]Album, 0, limit)
	for _, similar := range response.Similar.Artists {
		if len(albums) >= limit {
			break
		}
		name := strings.TrimSpace(similar.Name)
		if name == "" || strings.EqualFold(name, artist) {
			continue
		}
		top, err := c.getTopAlbums(ctx, name, 1)
		if err != nil || len(top) == 0 {
			continue
		}
		key := strings.ToLower(top[0].Artist + "\x00" + top[0].Title)
		if seen[key] {
			continue
		}
		seen[key] = true
		albums = append(albums, top[0])
	}
	if len(albums) == 0 {
		return nil, fmt.Errorf("no similar album recommendations found for %s", artist)
	}
	return albums, nil
}

// GetRecentTracks returns recent scrobbles for duplicate protection.
func (c *client) GetRecentTracks(ctx context.Context, user string, from time.Time) ([]RecentTrack, error) {
	user = strings.TrimSpace(user)
	if user == "" {
		return nil, errors.New("last.fm username is empty")
	}
	params := map[string]string{
		"method": "user.getRecentTracks",
		"user":   user,
		"limit":  "200",
	}
	if !from.IsZero() {
		params["from"] = strconv.FormatInt(from.Unix(), 10)
	} else {
		params["limit"] = "1"
	}
	var response struct {
		Recent struct {
			Tracks []struct {
				Name   string    `json:"name"`
				Artist textValue `json:"artist"`
				Album  textValue `json:"album"`
				Date   struct {
					UTS string `json:"uts"`
				} `json:"date"`
				Attr struct {
					NowPlaying string `json:"nowplaying"`
				} `json:"@attr"`
			} `json:"track"`
		} `json:"recenttracks"`
	}
	if err := c.doRequest(ctx, params, &response); err != nil {
		return nil, err
	}
	tracks := make([]RecentTrack, 0, len(response.Recent.Tracks))
	for _, item := range response.Recent.Tracks {
		artistName, title := item.Artist.String(), strings.TrimSpace(item.Name)
		if artistName == "" || title == "" {
			continue
		}
		var played time.Time
		if seconds, err := strconv.ParseInt(strings.TrimSpace(item.Date.UTS), 10, 64); err == nil && seconds > 0 {
			played = time.Unix(seconds, 0)
		}
		tracks = append(tracks, RecentTrack{Artist: artistName, Title: title, Album: item.Album.String(), Played: played, NowPlaying: item.Attr.NowPlaying == "1"})
	}
	return tracks, nil
}

func (c *client) Scrobble(ctx context.Context, artist, track, album string, timestamp int64) error {
	artist = strings.TrimSpace(artist)
	track = strings.TrimSpace(track)
	album = strings.TrimSpace(album)
	if artist == "" || track == "" {
		return errors.New("cannot scrobble with an empty artist or track")
	}
	if c.sessionKey == "" {
		return errors.New("last.fm session is not authenticated")
	}
	params := map[string]string{
		"method":    "track.scrobble",
		"artist":    artist,
		"track":     track,
		"timestamp": strconv.FormatInt(timestamp, 10),
		"album":     album,
		"sk":        c.sessionKey,
	}
	var response struct {
		Scrobbles struct {
			Attr struct {
				Accepted intValue `json:"accepted"`
				Ignored  intValue `json:"ignored"`
			} `json:"@attr"`
		} `json:"scrobbles"`
	}
	if err := c.doRequest(ctx, params, &response); err != nil {
		return err
	}
	if int(response.Scrobbles.Attr.Accepted) < 1 {
		return fmt.Errorf("last.fm rejected the scrobble (accepted=%d, ignored=%d)", response.Scrobbles.Attr.Accepted, response.Scrobbles.Attr.Ignored)
	}
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

// textValue accepts the inconsistent Last.fm JSON forms used for artist:
// either "Artist Name" or {"name":"Artist Name"}.
type textValue string

func (value *textValue) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*value = ""
		return nil
	}
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		*value = textValue(text)
		return nil
	}
	var object struct {
		Name string `json:"name"`
		Text string `json:"#text"`
	}
	if err := json.Unmarshal(data, &object); err != nil {
		return err
	}
	*value = textValue(firstNonEmpty(object.Name, object.Text))
	return nil
}
func (value textValue) String() string { return strings.TrimSpace(string(value)) }

type intValue int

func (value *intValue) UnmarshalJSON(data []byte) error {
	var number int
	if err := json.Unmarshal(data, &number); err == nil {
		*value = intValue(number)
		return nil
	}
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}
	number, err := strconv.Atoi(strings.TrimSpace(text))
	if err != nil && strings.TrimSpace(text) != "" {
		return err
	}
	*value = intValue(number)
	return nil
}

type apiTrack struct {
	Name     string   `json:"name"`
	Duration intValue `json:"duration"`
}

type trackValues []apiTrack

func (values *trackValues) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == "" {
		*values = nil
		return nil
	}
	var list []apiTrack
	if err := json.Unmarshal(data, &list); err == nil {
		*values = list
		return nil
	}
	var single apiTrack
	if err := json.Unmarshal(data, &single); err != nil {
		return err
	}
	*values = []apiTrack{single}
	return nil
}

type floatString float64

func (value *floatString) UnmarshalJSON(data []byte) error {
	var number float64
	if err := json.Unmarshal(data, &number); err == nil {
		*value = floatString(number)
		return nil
	}
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}
	number, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
	if err != nil && strings.TrimSpace(text) != "" {
		return err
	}
	*value = floatString(number)
	return nil
}

// deterministicParams is used only by tests/debugging when inspecting signed
// requests; keeping it here avoids map-order assumptions.
func deterministicParams(values url.Values) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
