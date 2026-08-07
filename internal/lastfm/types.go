package lastfm

import "time"

type Track struct {
	Title    string
	Duration int // seconds, 0 if unknown
}

type Album struct {
	Artist string
	Title  string
	Tracks []Track
}

type SimilarArtist struct {
	Name  string
	Match float64
}

type RecentTrack struct {
	Artist string
	Title  string
	Album  string
	Played time.Time
}
