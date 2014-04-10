package core

import (
	"log"

	"github.com/mdlayher/wavepipe/core/models"

	"github.com/wtolson/go-taglib"
)

// fsManager scans for media files in a specified path, and queues them up for inclusion
// in the wavepipe database
func fsManager(killFSChan chan struct{}) {
	log.Println("fs: starting...")

	// For now, this file just tests taglib
	file, err := taglib.Read("/tmp/test.flac")
	defer file.Close()
	if err != nil {
		log.Println(err)
		return
	}

	// Generate a song model from the file
	song, err := models.SongFromFile(file)
	if err != nil {
		log.Println(err)
		return
	}

	// Print tags
	log.Printf("%#v", song)
}
