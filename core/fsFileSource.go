package core

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mdlayher/wavepipe/data"

	"github.com/wtolson/go-taglib"
)

// fsFileSource represents a file source which indexes files in the local filesystem
type fsFileSource struct{}

// MediaScan scans for media files in the local filesystem
func (fsFileSource) MediaScan(mediaFolder string, verbose bool, walkCancelChan chan struct{}) error {
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
	songUpdateCount := 0
	folderCount := 0
	startTime := time.Now()

	// Cache entries which have been seen previously, to reduce database load
	folderCache := map[string]*data.Folder{}
	artistCache := map[string]*data.Artist{}
	albumCache := map[string]*data.Album{}

	if verbose {
		log.Println("fs: beginning media scan:", mediaFolder)
	} else {
		log.Println("fs: scanning:", mediaFolder)
	}

	// Invoke a recursive file walk on the given media folder, passing closure variables into
	// walkFunc to enable additional functionality
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

		// Check for a cached folder, or attempt to load it
		if tempFolder, ok := folderCache[folder.Path]; ok {
			folder = tempFolder
		} else if err := folder.Load(); err != nil && err == sql.ErrNoRows {
			// Make sure items actually exist at this path
			files, err := ioutil.ReadDir(folder.Path)
			if err != nil {
				log.Println(err)
				return err
			}

			// No items, skip it
			if len(files) == 0 {
				return nil
			}

			// Set short title
			folder.Title = path.Base(folder.Path)

			// Check for a parent folder
			pFolder := new(data.Folder)

			// If scan is triggered by a file, we have to check the dir twice to get parent
			if info.IsDir() {
				pFolder.Path = path.Dir(currPath)
			} else {
				pFolder.Path = path.Dir(path.Dir(currPath))
			}

			// Load parent
			if err := pFolder.Load(); err != nil && err != sql.ErrNoRows {
				log.Println(err)
				return err
			}

			// Copy parent folder's ID
			folder.ParentID = pFolder.ID

			// Save new folder
			if err := folder.Save(); err != nil {
				log.Println(err)
				return err
			}

			// Continue traversal
			folderCount++
		}

		// Cache this folder
		folderCache[folder.Path] = folder

		// Check for a valid media extension
		if !validSet.Has(path.Ext(currPath)) {
			return nil
		}

		// Attempt to scan media file with taglib
		file, err := taglib.Read(currPath)
		if err != nil {
			log.Println(err)
			return fmt.Errorf("%s: %s", currPath, err.Error())
		}
		defer file.Close()

		// Generate a song model from the TagLib file
		song, err := data.SongFromFile(file)
		if err != nil {
			log.Println(err)
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
		if tempArtist, ok := artistCache[artist.Title]; ok {
			artist = tempArtist
		} else if err := artist.Load(); err == sql.ErrNoRows {
			// Save new artist
			if err := artist.Save(); err != nil {
				log.Println(err)
			} else if err == nil {
				log.Printf("Artist: [#%05d] %s", artist.ID, artist.Title)
				artistCount++
			}
		}

		// Cache this artist
		artistCache[artist.Title] = artist

		// Generate the album model from this song's metadata
		album := data.AlbumFromSong(song)
		album.ArtistID = artist.ID

		// Generate cache key
		albumCacheKey := strconv.Itoa(album.ArtistID) + "_" + album.Title

		// Check for existing album
		// Note: if the album exists, this operation also loads necessary scanning information
		// such as the album ID, for use in song generation
		if tempAlbum, ok := albumCache[albumCacheKey]; ok {
			album = tempAlbum
		} else if err := album.Load(); err == sql.ErrNoRows {
			// Save album
			if err := album.Save(); err != nil {
				log.Println(err)
			} else if err == nil {
				log.Printf("  - Album: [#%05d] %s - %d - %s", album.ID, album.Artist, album.Year, album.Title)
				albumCount++
			}
		}

		// Cache this album
		albumCache[albumCacheKey] = album

		// Add ID fields to song
		song.ArtistID = artist.ID
		song.AlbumID = album.ID

		// Make a duplicate song to check if song has been modified since last scan
		song2 := new(data.Song)
		song2.FileName = song.FileName

		// Check for existing song
		if err := song2.Load(); err == sql.ErrNoRows {
			// Save song (don't log these because they really slow things down)
			if err2 := song.Save(); err2 != nil && er2r != sql.ErrNoRows {
				log.Println(err2)
			} else if err2 == nil {
				songCount++
			}
		} else {
			// Song already existed.  Check if it's been updated
			if song.LastModified > song2.LastModified {
				// Update existing song
				song.ID = song2.ID
				if err2 := song.Update(); err2 != nil {
					log.Println(err2)
				}

				songUpdateCount++
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
	if verbose {
		log.Printf("fs: media scan complete [time: %s]", time.Since(startTime).String())
		log.Printf("fs: added: [artists: %d] [albums: %d] [songs: %d] [folders: %d]", artistCount, albumCount, songCount, folderCount)
		log.Printf("fs: updated: [songs: %d]", songUpdateCount)
	}

	// No errors
	return nil
}

// OrphanScan scans for missing "orphaned" media files in the local filesystem
func (fsFileSource) OrphanScan(baseFolder string, subFolder string, verbose bool, orphanCancelChan chan struct{}) error {
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
	folderCount := 0
	songCount := 0
	startTime := time.Now()

	// Check if a baseFolder is set, meaning remove ANYTHING not under this base
	if baseFolder != "" {
		if verbose {
			log.Println("fs: orphan scanning base folder:", baseFolder)
		}

		// Scan for all songs NOT under the base folder
		songs, err := data.DB.SongsNotInPath(baseFolder)
		if err != nil {
			log.Println(err)
			return err
		}

		// Remove all songs which are not in this path
		for _, s := range songs {
			// Remove song from database
			if err := s.Delete(); err != nil {
				log.Println(err)
				return err
			}

			songCount++
		}

		// Scan for all folders NOT under the base folder
		folders, err := data.DB.FoldersNotInPath(baseFolder)
		if err != nil {
			log.Println(err)
			return err
		}

		// Remove all folders which are not in this path
		for _, f := range folders {
			// Remove folder from database
			if err := f.Delete(); err != nil {
				log.Println(err)
				return err
			}

			folderCount++
		}
	}

	// If no subfolder set, use the base folder to check file existence
	if subFolder == "" {
		subFolder = baseFolder
	}

	if verbose {
		log.Println("fs: orphan scanning subfolder:", subFolder)
	} else {
		log.Println("fs: removing:", subFolder)
	}

	// Scan for all songs in subfolder
	songs, err := data.DB.SongsInPath(subFolder)
	if err != nil {
		log.Println(err)
		return err
	}

	// Iterate all songs in this path
	for _, s := range songs {
		// Check that the song still exists in this place
		if _, err := os.Stat(s.FileName); os.IsNotExist(err) {
			// Remove song from database
			if err := s.Delete(); err != nil {
				log.Println(err)
				return err
			}

			songCount++
		}
	}

	// Scan for all folders in subfolder
	folders, err := data.DB.FoldersInPath(subFolder)
	if err != nil {
		return err
	}

	// Iterate all folders in this path
	for _, f := range folders {
		// Check that the folder still has items within it
		files, err := ioutil.ReadDir(f.Path)
		if err != nil && !os.IsNotExist(err) {
			log.Println(err)
			return err
		}

		// Delete any folders with 0 items
		if len(files) == 0 {
			if err := f.Delete(); err != nil {
				log.Println(err)
				return err
			}

			folderCount++
		}
	}

	// Now that songs have been purged, check for albums
	albumCount, err := data.DB.PurgeOrphanAlbums()
	if err != nil {
		log.Println(err)
		return err
	}

	// Check for artists
	artistCount, err := data.DB.PurgeOrphanArtists()
	if err != nil {
		log.Println(err)
		return err
	}

	// Print metrics
	if verbose {
		log.Printf("fs: orphan scan complete [time: %s]", time.Since(startTime).String())
		log.Printf("fs: removed: [artists: %d] [albums: %d] [songs: %d] [folders: %d]", artistCount, albumCount, songCount, folderCount)
	}
	return nil
}
