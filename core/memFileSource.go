package core

import (
	"database/sql"
	"log"
	"time"

	"github.com/mdlayher/wavepipe/data"
)

// memFileSource represents a file source which indexes mock files from memory
type memFileSource struct{}

// mockFiles is a slice of mock files to index
var mockFiles = []data.Song{
	data.Song{
		Album:        "Album",
		Artist:       "Artist",
		Bitrate:      320,
		Channels:     2,
		FileName:     "/mem/artist - song.mp3",
		FileSize:     1000,
		FileType:     "MP3",
		FolderID:     1,
		LastModified: time.Now().Unix(),
		Length:       60,
		SampleRate:   44100,
		Title:        "song",
		Track:        1,
		Year:         2014,
	},
	data.Song{
		Album:        "Album2",
		Artist:       "Artist2",
		Bitrate:      320,
		Channels:     2,
		FileName:     "/mem/artist2 - song2.mp3",
		FileSize:     1000,
		FileType:     "MP3",
		FolderID:     1,
		LastModified: time.Now().Unix(),
		Length:       60,
		SampleRate:   44100,
		Title:        "song2",
		Track:        1,
		Year:         2014,
	},
	data.Song{
		Album:        "Album3",
		Artist:       "Artist3",
		Bitrate:      320,
		Channels:     2,
		FileName:     "/mem/artist3 - song3.mp3",
		FileSize:     1000,
		FileType:     "MP3",
		FolderID:     1,
		LastModified: time.Now().Unix(),
		Length:       60,
		SampleRate:   44100,
		Title:        "song3",
		Track:        1,
		Year:         2014,
	},
}

// MediaScan adds mock media files to the database from memory
func (memFileSource) MediaScan(mediaFolder string, verbose bool, walkCancelChan chan struct{}) error {
	log.Println("mem: beginning mock media scan:", mediaFolder)

	// Iterate all media files and check for the matching prefix
	for _, song := range mockFiles {
		// Grab files with matching prefix
		if mediaFolder == song.FileName[0:len(mediaFolder)] {
			// Generate a folder model
			folder := new(data.Folder)
			folder.Path = "/mem"
			folder.Title = "/mem"

			// Attempt to load folder
			if err := folder.Load(); err == sql.ErrNoRows {
				// Save new folder
				if err := folder.Save(); err != nil {
					log.Println(err)
				}
			}

			// Generate an artist model from this song's metadata
			artist := data.ArtistFromSong(&song)

			// Check for existing artist
			// Note: if the artist exists, this operation also loads necessary scanning information
			// such as their artist ID, for use in album and song generation
			if err := artist.Load(); err == sql.ErrNoRows {
				// Save new artist
				if err := artist.Save(); err != nil {
					log.Println(err)
				}
			}

			// Generate the album model from this song's metadata
			album := data.AlbumFromSong(&song)
			album.ArtistID = artist.ID

			// Check for existing album
			// Note: if the album exists, this operation also loads necessary scanning information
			// such as the album ID, for use in song generation
			if err := album.Load(); err == sql.ErrNoRows {
				// Save album
				if err := album.Save(); err != nil {
					log.Println(err)
				}
			}

			// Add ID fields to song
			song.ArtistID = artist.ID
			song.AlbumID = album.ID

			// Check for existing song
			if err := song.Load(); err == sql.ErrNoRows {
				// Save song (don't log these because they really slow things down)
				if err := song.Save(); err != nil {
					log.Println(err)
				}
			}
		}
	}

	log.Println("mem: mock media scan complete")
	return nil
}

// OrphanScan does nothing for mock media files, because the database is temporary anyway
func (memFileSource) OrphanScan(baseFolder string, subFolder string, verbose bool, orphanCancelChan chan struct{}) error {
	return nil
}
