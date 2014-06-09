package common

import (
	"github.com/mdlayher/wavepipe/data"
)

// Metrics represents a variety of metrics about wavepipe's database
type Metrics struct {
	Artists int64 `json:"artists"`
	Albums  int64 `json:"albums"`
	Songs   int64 `json:"songs"`
	Folders int64 `json:"folders"`
}

// ServerMetrics returns a variety of metrics about the wavepipe database and server
func ServerMetrics() (*Metrics, error) {
	// Fetch total artists
	artists, err := data.DB.CountArtists()
	if err != nil {
		return nil, err
	}

	// Fetch total albums
	albums, err := data.DB.CountAlbums()
	if err != nil {
		return nil, err
	}

	// Fetch total songs
	songs, err := data.DB.CountSongs()
	if err != nil {
		return nil, err
	}

	// Fetch total folders
	folders, err := data.DB.CountFolders()
	if err != nil {
		return nil, err
	}

	// Combine all metrics
	return &Metrics{
		Artists: artists,
		Albums:  albums,
		Songs:   songs,
		Folders: folders,
	}, nil
}
