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

	"github.com/mdlayher/wavepipe/data"

	"github.com/wtolson/go-taglib"
)

// fsFileSource represents a file source which indexes files in the local filesystem
type fsFileSource struct{}

// MediaScan scans for media files in the local filesystem
func (fsFileSource) MediaScan(mediaFolder string, walkCancelChan chan struct{}) error {
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
	folderCount := 0
	startTime := time.Now()

	// Invoke a recursive file walk on the given media folder, passing closure variables into
	// walkFunc to enable additional functionality
	log.Println("fs: beginning media scan:", mediaFolder)
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

		// Check for an existing folder for this item
		folder := new(data.Folder)
		if info.IsDir() {
			// If directory, use this path
			folder.Path = currPath
		} else {
			// If file, use the directory path
			folder.Path = path.Dir(currPath)
		}

		// Attempt to load folder
		if err := folder.Load(); err != nil && err == sql.ErrNoRows {
			// Ensure this is actually a folder
			if !info.IsDir() {
				return nil
			}

			// Set short title
			folder.Title = path.Base(folder.Path)

			// Check for a parent folder
			pFolder := new(data.Folder)
			pFolder.Path = path.Dir(currPath)
			if err := pFolder.Load(); err != nil && err != sql.ErrNoRows {
				return err
			}

			// Copy parent folder's ID
			folder.ParentID = pFolder.ID

			// Save new folder
			if err := folder.Save(); err != nil {
				return err
			}

			// Continue traversal
			folderCount++
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
		song, err := data.SongFromFile(file)
		if err != nil {
			return err
		}

		// Populate filesystem-related struct fields using OS info
		song.FileName = currPath
		song.FileSize = info.Size()

		// Use this folder's ID
		song.FolderID = folder.ID

		// Extract type from the extension, capitalize it, drop the dot
		song.FileType = strings.ToUpper(path.Ext(info.Name()))[1:]
		song.LastModified = info.ModTime().Unix()

		// Generate an artist model from this song's metadata
		artist := data.ArtistFromSong(song)

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
		album := data.AlbumFromSong(song)
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
	log.Printf("fs: added: [artists: %d] [albums: %d] [songs: %d] [folders: %d]", artistCount, albumCount, songCount, folderCount)

	// No errors
	return nil
}

// OrphanScan scans for missing "orphaned" media files in the local filesystem
func (fsFileSource) OrphanScan(baseFolder string, subFolder string, orphanCancelChan chan struct{}) error {
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
		songs, err := data.DB.SongsNotInPath(baseFolder)
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
	songs, err := data.DB.SongsInPath(subFolder)
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
	albumCount, err := data.DB.PurgeOrphanAlbums()
	if err != nil {
		return err
	}

	// Finally, check for artists
	artistCount, err := data.DB.PurgeOrphanArtists()
	if err != nil {
		return err
	}

	// Print metrics
	log.Printf("fs: orphan scan complete [time: %s]", time.Since(startTime).String())
	log.Printf("fs: removed: [artists: %d] [albums: %d] [songs: %d]", artistCount, albumCount, songCount)
	return nil
}
