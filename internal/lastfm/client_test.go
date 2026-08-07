package lastfm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func testClient(server *httptest.Server) *client {
	return &client{
		apiKey:     "key",
		apiSecret:  "secret",
		username:   "listener",
		password:   "plain-password",
		baseURL:    server.URL,
		httpClient: server.Client(),
	}
}

func TestSearchAlbumsAcceptsNameAndStringArtist(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("method"); got != "album.search" {
			t.Fatalf("method = %q", got)
		}
		_, _ = w.Write([]byte(`{"results":{"albummatches":{"album":[{"name":"Horrified","artist":"Repulsion"}]}}}`))
	}))
	defer server.Close()

	albums, err := testClient(server).SearchAlbums(context.Background(), "Horrified")
	if err != nil {
		t.Fatal(err)
	}
	if len(albums) != 1 || albums[0].Artist != "Repulsion" || albums[0].Title != "Horrified" {
		t.Fatalf("albums = %#v", albums)
	}
}

func TestGetDiscographyUsesNameAndArtistFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"topalbums":{"album":[{"name":"Epidemic of Violence","artist":{"name":"Demolition Hammer"}},{"name":"Tortured Existence"}]}}`))
	}))
	defer server.Close()

	albums, err := testClient(server).GetDiscography(context.Background(), "Demolition Hammer")
	if err != nil {
		t.Fatal(err)
	}
	if len(albums) != 2 {
		t.Fatalf("albums = %#v", albums)
	}
	for _, album := range albums {
		if album.Artist != "Demolition Hammer" || album.Title == "" {
			t.Fatalf("album = %#v", album)
		}
	}
}

func TestGetAlbumTracksAcceptsStringArtistAndSeconds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"album":{"name":"Epidemic of Violence","artist":"Demolition Hammer","tracks":{"track":[{"name":"Skull Fracturing Nightmare","duration":"240"}]}}}`))
	}))
	defer server.Close()

	album, err := testClient(server).GetAlbumTracks(context.Background(), "Demolition Hammer", "Epidemic of Violence")
	if err != nil {
		t.Fatal(err)
	}
	if album.Artist != "Demolition Hammer" || album.Title != "Epidemic of Violence" {
		t.Fatalf("album = %#v", album)
	}
	if len(album.Tracks) != 1 || album.Tracks[0].Duration != 240 {
		t.Fatalf("tracks = %#v", album.Tracks)
	}
}

func TestAuthenticateSendsPlainPassword(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if got := r.Form.Get("password"); got != "plain-password" {
			t.Fatalf("password = %q", got)
		}
		_, _ = w.Write([]byte(`{"session":{"key":"session-key"}}`))
	}))
	defer server.Close()

	client := testClient(server)
	if err := client.Authenticate(context.Background()); err != nil {
		t.Fatal(err)
	}
	if client.sessionKey != "session-key" {
		t.Fatalf("sessionKey = %q", client.sessionKey)
	}
}

func TestScrobbleReadsJSONAttr(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		required := url.Values{
			"artist":    {"Demolition Hammer"},
			"track":     {"Skull Fracturing Nightmare"},
			"album":     {"Epidemic of Violence"},
			"timestamp": {"1700000000"},
			"sk":        {"session-key"},
		}
		for key, values := range required {
			if got := r.Form.Get(key); got != values[0] {
				t.Fatalf("%s = %q", key, got)
			}
		}
		_, _ = w.Write([]byte(`{"scrobbles":{"@attr":{"accepted":1,"ignored":0}}}`))
	}))
	defer server.Close()

	client := testClient(server)
	client.sessionKey = "session-key"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := client.Scrobble(ctx, "Demolition Hammer", "Skull Fracturing Nightmare", "Epidemic of Violence", 1700000000); err != nil {
		t.Fatal(err)
	}
}

func TestAuthenticateUsesConfiguredSessionKeyWithoutPassword(t *testing.T) {
	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		t.Fatal("Authenticate should not make a request when a session key is configured")
	}))
	defer server.Close()

	client := testClient(server)
	client.password = ""
	client.sessionKey = "saved-session-key"
	if err := client.Authenticate(context.Background()); err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("unexpected authentication request")
	}
}

func TestGetSimilarAlbumsUsesSimilarArtistsTopAlbum(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("method") {
		case "artist.getSimilar":
			_, _ = w.Write([]byte(`{"similarartists":{"artist":[{"name":"Kreator","match":"0.9"},{"name":"Sodom","match":"0.8"}]}}`))
		case "artist.getTopAlbums":
			artist := r.URL.Query().Get("artist")
			_, _ = w.Write([]byte(`{"topalbums":{"album":[{"name":"Top Album","artist":{"name":"` + artist + `"}}]}}`))
		default:
			t.Fatalf("unexpected method %q", r.URL.Query().Get("method"))
		}
	}))
	defer server.Close()

	albums, err := testClient(server).GetSimilarAlbums(context.Background(), "Demolition Hammer", 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(albums) != 2 || albums[0].Artist != "Kreator" || albums[1].Artist != "Sodom" {
		t.Fatalf("albums = %#v", albums)
	}
}

func TestGetRecentTracksParsesStringValuesAndFrom(t *testing.T) {
	from := time.Unix(1700000000, 0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("from"); got != "1700000000" {
			t.Fatalf("from = %q", got)
		}
		_, _ = w.Write([]byte(`{"recenttracks":{"track":[{"name":"Hell Awaits","artist":"Slayer","album":"Hell Awaits","date":{"uts":"1700000010"}}]}}`))
	}))
	defer server.Close()

	tracks, err := testClient(server).GetRecentTracks(context.Background(), "deathrashed", from)
	if err != nil {
		t.Fatal(err)
	}
	if len(tracks) != 1 || tracks[0].Artist != "Slayer" || tracks[0].Title != "Hell Awaits" || tracks[0].Played.Unix() != 1700000010 {
		t.Fatalf("tracks = %#v", tracks)
	}
}
