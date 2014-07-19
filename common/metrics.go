package common

import (
	"github.com/mdlayher/wavepipe/data"
)

// Metrics represents a variety of metrics about the current wavepipe instance, and contains several
// nested structs which contain more specific metrics
type Metrics struct {
	Database *DatabaseMetrics `json:"database"`
}

// DatabaseMetrics represents metrics regarding the wavepipe database, including total numbers
// of specific objects, and the time when the database was last updated
type DatabaseMetrics struct {
	Updated int64 `json:"updated"`

	Artists int64 `json:"artists"`
	Albums  int64 `json:"albums"`
	Songs   int64 `json:"songs"`
	Folders int64 `json:"folders"`
	Art     int64 `json:"art"`
}

// GetDatabaseMetrics returns a variety of metrics about the wavepipe database, including
// total numbers of specific objects, and the time when the database was last updated
func GetDatabaseMetrics() (*DatabaseMetrics, error) {
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

	// Fetch total art
	art, err := data.DB.CountArt()
	if err != nil {
		return nil, err
	}

	// Combine all metrics
	return &DatabaseMetrics{
		Updated: ScanTime(),
		Artists: artists,
		Albums:  albums,
		Songs:   songs,
		Art:     art,
		Folders: folders,
	}, nil
}
