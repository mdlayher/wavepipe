package core

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mdlayher/goset"
	"github.com/wtolson/go-taglib"
)

// validSet is a set of valid file extensions which we should scan as media, as they are the ones
// which TagLib is capable of reading
var validSet = set.New(".ape", ".flac", ".m4a", ".mp3", ".mpc", ".ogg", ".wma", ".wv")

// fsTask is the interface which defines a filesystem task, such as a media scan, or an orphan scan
type fsTask interface {
	Scan(string, string, chan struct{}) error
}

// fsManager handles fsWalker processes, and communicates back and forth with the manager goroutine
func fsManager(mediaFolder string, fsKillChan chan struct{}) {
	log.Println("fs: starting...")

	// Initialize a queue of filesystem tasks
	fsQueue := make(chan fsTask, 10)
	cancelQueue := make(chan chan struct{}, 10)

	// Queue an orphan scan, followed by a media scan
	fsQueue <- new(fsOrphanScan)
	fsQueue <- new(fsMediaScan)

	// Invoke task queue via goroutine, so it can be halted via the manager
	go func() {
		for {
			select {
			// Trigger a fsTask from queue
			case task := <-fsQueue:
				// Create a channel to halt the scan
				cancelChan := make(chan struct{})
				cancelQueue <- cancelChan

				// Start the scan
				// TODO: use actual base folder, actual subfolders from filesystem watcher
				if err := task.Scan(mediaFolder, "", cancelChan); err != nil {
					log.Println(err)
				}

				// On completion, close the cancel channel
				cancelChan = <-cancelQueue
				close(cancelChan)
			}
		}
	}()

	// Trigger manager events via channel
	for {
		select {
		// Stop filesystem manager
		case <-fsKillChan:
			// Halt any in-progress tasks
			log.Println("fs: halting tasks")
			for i := 0; i < len(cancelQueue); i++ {
				// Receive a channel
				f := <-cancelQueue
				if f == nil {
					continue
				}

				// Send termination
				f <- struct{}{}
				log.Println("fs: task halted")
			}

			// Inform manager that shutdown is complete
			log.Println("fs: stopped!")
			fsKillChan <- struct{}{}
			return
		}
	}
}

// fsMediaScan represents a filesystem task which scans the given path for new media
type fsMediaScan struct{}

// Scan scans for media files in a specified path, and queues them up for inclusion
// in the wavepipe database
func (fs *fsMediaScan) Scan(mediaFolder string, subFolder string, walkCancelChan chan struct{}) error {
	// Media scans are comprehensive, so subfolder has no purpose
	if subFolder != "" {
		return errors.New("media scan: subfolder not valid for media scan operation")
	}

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

	// Invoke a recursive file walk on the given media folder, passing closure variables into
	// walkFunc to enable additional functionality
	log.Println("fs: beginning media scan")
	err := filepath.Walk(mediaFolder, func(currPath string, info os.FileInfo, err error) error {
		// Stop walking immediately if needed
		mutex.RLock()
		if haltWalk {
			return errors.New("media scan: halted by channel")
		}
		mutex.RUnlock()

		// Make sure path is actually valid
		if info == nil {
			return errors.New("media scan: invalid path: " + currPath)
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

		// Generate a song model from the TagLib file
		song, err := SongFromFile(file)
		if err != nil {
			return err
		}

		// Populate filesystem-related struct fields using OS info
		song.FileName = currPath
		song.FileSize = info.Size()

		// Extract type from the extension, capitalize it, drop the dot
		song.FileType = strings.ToUpper(path.Ext(info.Name()))[1:]
		song.LastModified = info.ModTime().Unix()

		// Generate an artist model from this song's metadata
		artist := ArtistFromSong(song)

		// Check for existing artist
		// Note: if the artist exists, this operation also loads necessary scanning information
		// such as their artist ID, for use in album and song generation
		if err := artist.Load(); err == sql.ErrNoRows {
			// Save new artist
			if err := artist.Save(); err != nil {
				log.Println(err)
			} else if err == nil {
				log.Printf("Artist: [%04d] %s", artist.ID, artist.Title)
				artistCount++
			}
		}

		// Generate the album model from this song's metadata
		album := AlbumFromSong(song)
		album.ArtistID = artist.ID

		// Check for existing album
		// Note: if the album exists, this operation also loads necessary scanning information
		// such as the album ID, for use in song generation
		if err := album.Load(); err == sql.ErrNoRows {
			// Save album
			if err := album.Save(); err != nil {
				log.Println(err)
			} else if err == nil {
				log.Printf("  - Album: [%04d] %s - %d - %s", album.ID, album.Artist, album.Year, album.Title)
				albumCount++
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
			} else if err == nil {
				songCount++
			}
		}

		// Successful media scan
		return nil
	})

	// Check for filesystem walk errors
	if err != nil {
		return err
	}

	// Print metrics
	log.Printf("fs: media scan complete [time: %s]", time.Since(startTime).String())
	log.Printf("fs: added: [artists: %d] [albums: %d] [songs: %d]", artistCount, albumCount, songCount)

	// No errors
	return nil
}

// fsOrphanScan represents a filesystem task which scans the given path for orphaned media
type fsOrphanScan struct{}

// Scan scans for media files which have been removed from the media directory, and removes
// them as appropriate.  An orphan is defined as follows:
//   - Artist: no more songs contain this artist's ID
//   - Album: no more songs contain this album's ID
//   - Song: song is no longer present in the filesystem
// The baseFolder is the root location of the media folder.  As wavepipe currently supports only
// one media folder, any media which does not reside in this folder is orphaned.  If left blank,
// only the subFolder will be checked.
// The subFolder is the current file location, under the baseFolder.  This is used to allow for
// quick scans of a small subsection of the directory, such as on a filesystem change.  Any files
// which are in the database, but do not exist on disk, will be orphaned and removed.
func (fs *fsOrphanScan) Scan(baseFolder string, subFolder string, orphanCancelChan chan struct{}) error {
	// If both folders are empty, there is nothing to do
	if baseFolder == "" && subFolder == "" {
		return errors.New("orphan scan: no base folder or subfolder")
	}

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
	songCount := 0
	startTime := time.Now()

	log.Println("fs: beginning orphan scan")

	// Check if a baseFolder is set, meaning remove ANYTHING not under this base
	if baseFolder != "" {
		log.Println("fs: orphan scanning base folder:", baseFolder)

		// Scan for all songs NOT under the base folder
		songs, err := db.SongsNotInPath(baseFolder)
		if err != nil {
			return err
		}

		// Remove all songs which are not in this path
		for _, s := range songs {
			// Remove song from database
			if err := s.Delete(); err != nil {
				return err
			}

			songCount++
		}
	}

	// If no subfolder set, use the base folder to check file existence
	if subFolder == "" {
		subFolder = baseFolder
	}

	// Scan for all songs in subfolder
	log.Println("fs: orphan scanning subfolder:", subFolder)
	songs, err := db.SongsInPath(subFolder)
	if err != nil {
		return err
	}

	// Iterate all songs in this path
	for _, s := range songs {
		// Check that the song still exists in this place
		if _, err := os.Stat(s.FileName); os.IsNotExist(err) {
			// Remove song from database
			if err := s.Delete(); err != nil {
				return err
			}

			songCount++
		}
	}

	// Now that songs have been purged, check for albums
	albumCount, err := db.PurgeOrphanAlbums()
	if err != nil {
		return nil
	}

	// Finally, check for artists
	artistCount, err := db.PurgeOrphanArtists()
	if err != nil {
		return nil
	}

	// Print metrics
	log.Printf("fs: orphan scan complete [time: %s]", time.Since(startTime).String())
	log.Printf("fs: removed: [artists: %d] [albums: %d] [songs: %d]", artistCount, albumCount, songCount)

	// No errors
	return nil
}
