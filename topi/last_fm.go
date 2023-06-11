package topi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type LastFM struct {
	Track *LastFMTrack
	URL   string
	Error string
}

type LastFMTrack struct {
	Name      string
	Artist    string
	ArtistURL string
	Album     string
	Artwork   string
	URL       string
	Loved     bool
}

type LastFMResponse struct {
	RecentTracks struct {
		Track []struct {
			Artist struct {
				URL   string `json:"url"`
				Name  string `json:"name"`
				Image Image  `json:"image"`
				MbID  string `json:"mbid"`
			} `json:"artist"`
			Date struct {
				Uts  int64  `json:"uts,string"`
				Text string `json:"#text"`
			} `json:"date"`
			MbID       string `json:"mbid"`
			Name       string `json:"name"`
			Image      Image  `json:"image"`
			URL        string `json:"url"`
			Streamable int    `json:"streamable,string"`
			Album      struct {
				MbID string `json:"mbid"`
				Text string `json:"#text"`
			}
			Loved int `json:"loved,string"`
			Attr  struct {
				NowPlaying string `json:"nowplaying"`
			} `json:"@attr"`
		}
	} `json:"recenttracks"`
	Attr struct {
		User       string `json:"user"`
		TotalPages int    `json:"totalPages,string"`
		Page       int    `json:"page,string"`
		Total      int    `json:"total,string"`
		PerPage    int    `json:"perPage,string"`
	} `json:"@attr"`
}

type Image []struct {
	Size string `json:"size"`
	Text string `json:"#text"`
}

func (s *Server) FetchLastFM(ctx context.Context) LastFM {
	url := fmt.Sprintf("https://ws.audioscrobbler.com/2.0/?method=%s&user=%s&api_key=%s&format=%s&limit=%d&extended=%d", "user.getrecenttracks", s.cfg.LastFM.Username, s.cfg.LastFM.APIKey, "json", 1, 1)
	rq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return LastFM{Error: fmt.Sprintf("failed to create request: %s", err)}
	}
	rs, err := s.httpClient.Do(rq)
	if err != nil {
		return LastFM{Error: fmt.Sprintf("failed to do request: %s", err)}
	}
	defer rs.Body.Close()

	var resp LastFMResponse
	if err = json.NewDecoder(rs.Body).Decode(&resp); err != nil {
		return LastFM{Error: fmt.Sprintf("failed to decode response: %s", err)}
	}

	var track *LastFMTrack
	if len(resp.RecentTracks.Track) > 0 {
		lastFmTrack := resp.RecentTracks.Track[0]
		if lastFmTrack.Attr.NowPlaying == "true" {
			track = &LastFMTrack{
				Name:      lastFmTrack.Name,
				Artist:    lastFmTrack.Artist.Name,
				ArtistURL: lastFmTrack.Artist.URL,
				Album:     lastFmTrack.Album.Text,
				Artwork:   lastFmTrack.Image[len(lastFmTrack.Image)-1].Text,
				URL:       lastFmTrack.URL,
			}
		}
	}

	return LastFM{
		Track: track,
		URL:   fmt.Sprintf("https://www.last.fm/user/%s", s.cfg.LastFM.Username),
	}
}
