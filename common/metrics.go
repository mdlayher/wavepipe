package common

import (
	"sync/atomic"

	"github.com/mdlayher/wavepipe/data"
)

var (
	// rxBytes is the total number of bytes received over the network
	rxBytes int64
	// txBytes is the total number of bytes received over the network
	txBytes int64
)

// Metrics represents a variety of metrics about the current wavepipe instance, and contains several
// nested structs which contain more specific metrics
type Metrics struct {
	Database *DatabaseMetrics `json:"database"`
	Network  *NetworkMetrics  `json:"network"`
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

// NetworkMetrics represents metrics regarding wavepipe network traffic, including total traffic
// received and transmitted in bytes
type NetworkMetrics struct {
	RXBytes int64 `json:"rxBytes"`
	TXBytes int64 `json:"txBytes"`
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

// AddRXBytes atomically increments the rxBytes counter by the amount specified
func AddRXBytes(count int64) {
	atomic.AddInt64(&rxBytes, count)
}

// AddTXBytes atomically increments the txBytes counter by the amount specified
func AddTXBytes(count int64) {
	atomic.AddInt64(&txBytes, count)
}

// RXBytes returns the total number of bytes received over the network
func RXBytes() int64 {
	return atomic.LoadInt64(&rxBytes)
}

// TXBytes returns the total number of bytes transmitted over the network
func TXBytes() int64 {
	return atomic.LoadInt64(&txBytes)
}
