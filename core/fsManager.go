package core

import (
	"log"

	"github.com/mdlayher/wavepipe/core/models"

	"github.com/vbatts/go-taglib/taglib"
)

// fsManager scans for media files in a specified path, and queues them up for inclusion
// in the wavepipe database
func fsManager(killFSChan chan struct{}) {
	log.Println("fs: starting...")

	// For now, this file just tests taglib
	file := taglib.Open("/tmp/test.flac")
	defer file.Close()

	// Generate a song model from the file
	song, err := models.SongFromFile(file)
	if err != nil {
		log.Println(err)
		return
	}

	// Print tags
	log.Printf("%#v", song)
}
