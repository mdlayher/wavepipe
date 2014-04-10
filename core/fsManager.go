package core

import (
	"errors"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/mdlayher/wavepipe/core/models"

	"github.com/mdlayher/goset"
	"github.com/wtolson/go-taglib"
)

// validSet is a set of valid file extensions which we should scan as media, as they are the ones
// which TagLib is capable of reading
var validSet = set.New(".ape", ".flac", ".m4a", ".mp3", ".mpc", ".ogg", ".wma", ".wv")

// fsManager scans for media files in a specified path, and queues them up for inclusion
// in the wavepipe database
func fsManager(mediaFolder string, fsKillChan chan struct{}) {
	log.Println("fs: starting...")

	// Keep sets of unique artists, albums, and songs encountered
	artistSet := set.New()
	albumSet := set.New()
	songSet := set.New()

	// Invoke a recursive file walk on the given media folder, passing closure variables into
	// walkFunc to enable additional functionality
	err := filepath.Walk(mediaFolder, func(currPath string, info os.FileInfo, err error) error {
		// Make sure path is actually valid
		if info == nil {
			return errors.New("walk: invalid path: " + currPath)
		}

		// Ignore directories for now
		if info.IsDir() {
			return nil
		}

		// Check for a valid media extension
		if !validSet.Has(path.Ext(currPath)) {
			return nil
		}

		// Attempt to scan media file with taglib
		file, err := taglib.Read(currPath)
		if err != nil {
			return err
		}
		defer file.Close()

		// Generate a song model from the file
		song, err := models.SongFromFile(file)
		if err != nil {
			return err
		}

		// Keep track of unique sets
		artistSet.Add(song.Artist)
		albumSet.Add(song.Album)
		songSet.Add(song)

		// Print tags
		log.Printf("%s - %s", song.Artist, song.Title)
		return nil
	})

	// Check for filesystem walk errors
	if err != nil {
		log.Println(err)
	}

	log.Println(artistSet)
	log.Println(albumSet)
	log.Println(songSet)

	// Trigger events via channel
	for {
		select {
		// Stop filesystem manager
		case <-fsKillChan:
			// Inform manager that shutodwn is complete
			log.Println("fs: stopped!")
			fsKillChan <- struct{}{}
			return
		}
	}
}

