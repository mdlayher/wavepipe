package core

import (
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/mdlayher/wavepipe/core/models"

	"github.com/mdlayher/goset"
	"github.com/wtolson/go-taglib"
)

// validSet is a set of valid file extensions which we should scan as media
var validSet = set.New(".flac", ".mp3")

// fsManager scans for media files in a specified path, and queues them up for inclusion
// in the wavepipe database
func fsManager(mediaFolder string, killFSChan chan struct{}) {
	log.Println("fs: starting...")

	// Invoke a recursive file walk on the given media folder
	err := filepath.Walk(mediaFolder, walkFn)
	if err != nil {
		log.Println(err)
		return
	}
}

// walkFn is called by filepath.Walk() to recursively traverse a directory structure,
// searching for media to include in the wavepipe database
func walkFn(currPath string, info os.FileInfo, err error) error {
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

	// Print tags
	log.Printf("%s - %s", song.Artist, song.Title)
	return nil
}
