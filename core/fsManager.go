package core

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/mdlayher/goset"
	"github.com/wtolson/go-taglib"
)

// validSet is a set of valid file extensions which we should scan as media, as they are the ones
// which TagLib is capable of reading
var validSet = set.New(".ape", ".flac", ".m4a", ".mp3", ".mpc", ".ogg", ".wma", ".wv")

// fsManager handles fsWalker processes, and communicates back and forth with the manager goroutine
func fsManager(mediaFolder string, fsKillChan chan struct{}) {
	log.Println("fs: starting...")

	// Trigger an orphan scan, which can be halted via channel
	orphanCancelChan := make(chan struct{})
	orphanErrChan := fsOrphanScan(mediaFolder, orphanCancelChan)

	// Trigger a filesystem walk, which can be halted via channel
	walkCancelChan := make(chan struct{})
	walkErrChan := fsWalker(mediaFolder, walkCancelChan)

	// Trigger events via channel
	for {
		select {
		// Stop filesystem manager
		case <-fsKillChan:
			// Halt any in-progress walks
			orphanCancelChan <- struct{}{}
			walkCancelChan <- struct{}{}

			// Inform manager that shutdown is complete
			log.Println("fs: stopped!")
			fsKillChan <- struct{}{}
			return
		// Filesystem orphan error return channel
		case err := <-orphanErrChan:
			// Check if error occurred
			if err == nil {
				break
			}

			// Report orphan errors
			log.Println(err)
		// Filesystem orphan error return channel
		case err := <-walkErrChan:
			// Check if error occurred
			if err == nil {
				break
			}

			// Report walk errors
			log.Println(err)
		}
	}
}

// fsWalker scans for media files in a specified path, and queues them up for inclusion
// in the wavepipe database
func fsWalker(mediaFolder string, walkCancelChan chan struct{}) chan error {
	// Return errors on channel
	errChan := make(chan error)

	// Halt walk if needed
	var mutex sync.RWMutex
	haltWalk := false
	go func() {
		// Wait for signal
		<-walkCancelChan

		// Halt!
		mutex.Lock()
		haltWalk = true
		mutex.Unlock()
	}()

	// Track metrics about the walk
	artistCount := 0
	albumCount := 0
	songCount := 0
	startTime := time.Now()

	// Invoke walker goroutine
	go func() {
		// Invoke a recursive file walk on the given media folder, passing closure variables into
		// walkFunc to enable additional functionality
		log.Println("fs: beginning file walk")
		err := filepath.Walk(mediaFolder, func(currPath string, info os.FileInfo, err error) error {
			// Stop walking immediately if needed
			mutex.RLock()
			if haltWalk {
				return errors.New("walk: halted by channel")
			}
			mutex.RUnlock()

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
				return fmt.Errorf("%s: %s", currPath, err.Error())
			}
			defer file.Close()

			// Generate a song model from the TagLib file, and the OS file
			song, err := SongFromFile(file, info)
			if err != nil {
				return err
			}

			// Generate an artist model from this song's metadata
			artist := ArtistFromSong(song)

			// Check for existing artist
			if err := artist.Load(); err == sql.ErrNoRows {
				// Save new artist
				if err := artist.Save(); err != nil {
					log.Println(err)
				} else if err == nil {
					log.Printf("New artist: [%02d] %s", artist.ID, artist.Title)
					artistCount++
				}
			}

			// Generate the album model from this song's metadata
			album := AlbumFromSong(song)
			album.ArtistID = artist.ID

			// Check for existing album
			if err := album.Load(); err == sql.ErrNoRows {
				// Save album
				if err := album.Save(); err != nil {
					log.Println(err)
				} else if err == nil {
					log.Printf("New album: [%02d] %s - %s", album.ID, album.Artist, album.Title)
					albumCount++
				}
			}

			// Add ID fields to song
			song.ArtistID = artist.ID
			song.AlbumID = album.ID

			// Check for existing song
			if err := song.Load(); err == sql.ErrNoRows {
				// Save song
				if err := song.Save(); err != nil {
					log.Println(err)
				} else if err == nil {
					log.Printf("New song: [%02d] %s - %s - %s", song.ID, song.Artist, song.Album, song.Title)
					songCount++
				}
			}

			return nil
		})

		// Check for filesystem walk errors
		if err != nil {
			errChan <- err
			return
		}

		// Print metrics
		log.Printf("fs: file walk complete [time: %s]", time.Since(startTime).String())
		log.Printf("fs: [artists: %d] [albums: %d] [songs: %d]", artistCount, albumCount, songCount)

		// No errors
		errChan <- nil
	}()

	// Return communication channel
	return errChan
}

// fsOrphanScan scans for media files which have been removed from the media directory, and removes
// them as appropriate.  An orphan is defined as follows:
//   - Artist: no more songs contain this artist's ID
//   - Album: no more songs contain this album's ID
//   - Song: song is no longer present in the filesystem
func fsOrphanScan(mediaFolder string, orphanCancelChan chan struct{}) chan error {
	// Return errors on channel
	errChan := make(chan error)

	// Halt scan if needed
	var mutex sync.RWMutex
	haltOrphanScan := false
	go func() {
		// Wait for signal
		<-orphanCancelChan

		// Halt!
		mutex.Lock()
		haltOrphanScan = true
		mutex.Unlock()
	}()

	// Track metrics about the scan
	artistCount := 0
	albumCount := 0
	songCount := 0
	startTime := time.Now()

	// Invoke scanner goroutine
	go func() {
		log.Println("fs: beginning orphan scan")

		// Print metrics
		log.Printf("fs: orphan scan complete [time: %s]", time.Since(startTime).String())
		log.Printf("fs: [artists: %d] [albums: %d] [songs: %d]", artistCount, albumCount, songCount)

		// No errors
		errChan <- nil
	}()

	// Return communication channel
	return errChan
}
