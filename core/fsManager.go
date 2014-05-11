package core

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/mdlayher/goset"
	"github.com/romanoff/fsmonitor"
)

// artSet is a set of valid file extensions which we should scan as art
var artSet = set.New(".jpg", ".jpeg", ".png")

// mediaSet is a set of valid file extensions which we should scan as media, as they are the ones
// which TagLib is capable of reading
var mediaSet = set.New(".ape", ".flac", ".m4a", ".mp3", ".mpc", ".ogg", ".wma", ".wv")

// fsQueue is a queue of tasks to be performed by the filesystem, such as media and orphan scans
var fsQueue = make(chan fsTask, 10)

// fsSource is the data source used to scan for media files (could be filesystem, memory, etc)
var fsSource fileSource

// fsTask is the interface which defines a filesystem task, such as a media, or orphan scan
type fsTask interface {
	Folders() (string, string)
	SetFolders(string, string)
	Scan(string, string, chan struct{}) error
	Verbose(bool)
}

// fileSource represents a source from which files can be scanned and indexed
type fileSource interface {
	MediaScan(string, bool, chan struct{}) error
	OrphanScan(string, string, bool, chan struct{}) error
}

// fsManager handles fsWalker processes, and communicates back and forth with the manager goroutine
func fsManager(mediaFolder string, fsKillChan chan struct{}) {
	log.Println("fs: starting...")

	// Initialize a queue to cancel filesystem tasks
	cancelQueue := make(chan chan struct{}, 10)

	// Set up the data source (typically filesystem, unless in test mode)
	fsSource = fsFileSource{}
	if os.Getenv("WAVEPIPE_TEST") == "1" {
		// Mock file source
		fsSource = memFileSource{}
	}

	// Track the number of filesystem events fired
	fsTaskCount := 0

	// Initialize filesystem watcher when ready
	watcherChan := make(chan struct{})

	// Queue an initial, verbose orphan scan
	o := new(fsOrphanScan)
	o.SetFolders(mediaFolder, "")
	o.Verbose(true)
	fsQueue <- o

	// Queue a media scan
	m := new(fsMediaScan)
	m.SetFolders(mediaFolder, "")
	m.Verbose(true)
	fsQueue <- m

	// Invoke task queue via goroutine, so it can be halted via the manager
	go func() {
		for {
			select {
			// Trigger a fsTask from queue
			case task := <-fsQueue:
				// Create a channel to halt the scan
				cancelChan := make(chan struct{})
				cancelQueue <- cancelChan

				// Retrieve the folders to use with scan
				baseFolder, subFolder := task.Folders()

				// Start the scan
				if err := task.Scan(baseFolder, subFolder, cancelChan); err != nil {
					log.Println(err)
				}

				// On completion, close the cancel channel
				cancelChan = <-cancelQueue
				close(cancelChan)
				fsTaskCount++

				// After both initial scans complete, start the filesystem watcher
				if fsTaskCount == 2 {
					close(watcherChan)
				}
			}
		}
	}()

	// Create a filesystem watcher, which is triggered after initial scans
	go func() {
		// Block until triggered
		<-watcherChan

		// Initialize the watcher
		watcher, err := fsmonitor.NewWatcher()
		if err != nil {
			log.Println(err)
			return
		}

		// Wait for events on goroutine
		go func() {
			// Recently modified/renamed files sets, used as rate-limiters to prevent modify
			// events from flooding the select statement.  The filesystem watcher may fire an
			// excessive number of events, so these will block the extras for a couple seconds.
			recentModifySet := set.New()
			recentRenameSet := set.New()

			for {
				select {
				// Event occurred
				case ev := <-watcher.Event:
					switch {
					// On modify, trigger a media scan
					case ev.IsModify():
						// Add file to set, stopping it from propogating if the event was recently triggered
						if !recentModifySet.Add(ev.Name) {
							break
						}

						// Remove file from rate-limiting set after a couple seconds
						go func() {
							<-time.After(2 * time.Second)
							recentModifySet.Remove(ev.Name)
						}()

						fallthrough
					// On create, trigger a media scan
					case ev.IsCreate():
						// Invoke a slight delay to enable file creation
						<-time.After(250 * time.Millisecond)

						// Scan item as the "base folder", so it just adds this item
						m := new(fsMediaScan)
						m.SetFolders(ev.Name, "")
						m.Verbose(false)
						fsQueue <- m
					// On rename, trigger an orphan scan
					case ev.IsRename():
						// Add file to set, stopping it from propogating if the event was recently triggered
						if !recentRenameSet.Add(ev.Name) {
							break
						}

						// Remove file from rate-limiting set after a couple seconds
						go func() {
							<-time.After(2 * time.Second)
							recentRenameSet.Remove(ev.Name)
						}()

						fallthrough
					// On delete, trigger an orphan scan
					case ev.IsDelete():
						// Invoke a slight delay to enable file deletion
						<-time.After(250 * time.Millisecond)

						// Scan item as the "subfolder", so it just removes this item
						o := new(fsOrphanScan)
						o.SetFolders("", ev.Name)
						o.Verbose(false)
						fsQueue <- o
					}
				// Watcher errors
				case err := <-watcher.Error:
					log.Println(err)
					return
				}
			}
		}()

		// Watch media folder
		if err := watcher.Watch(mediaFolder); err != nil {
			log.Println(err)
			return
		}
		log.Println("fs: watching folder:", mediaFolder)
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
type fsMediaScan struct {
	baseFolder string
	subFolder  string
	verbose    bool
}

// Folders returns the base folder and subfolder for use with a scanning task
func (fs *fsMediaScan) Folders() (string, string) {
	return fs.baseFolder, fs.subFolder
}

// SetFolders sets the base folder and subfolder for use with a scanning task
func (fs *fsMediaScan) SetFolders(baseFolder string, subFolder string) {
	fs.baseFolder = baseFolder
	fs.subFolder = subFolder
}

// Verbose sets the verbosity level of the scanning task
func (fs *fsMediaScan) Verbose(verbose bool) {
	fs.verbose = verbose
}

// Scan scans for media files in a specified path, and queues them up for inclusion
// in the wavepipe database
func (fs *fsMediaScan) Scan(mediaFolder string, subFolder string, walkCancelChan chan struct{}) error {
	// Media scans are comprehensive, so subfolder has no purpose
	if subFolder != "" {
		return errors.New("media scan: subfolder not valid for media scan operation")
	}

	// Scan for media using the specified file source
	return fsSource.MediaScan(mediaFolder, fs.verbose, walkCancelChan)
}

// fsOrphanScan represents a filesystem task which scans the given path for orphaned media
type fsOrphanScan struct {
	baseFolder string
	subFolder  string
	verbose    bool
}

// Folders returns the base folder and subfolder for use with a scanning task
func (fs *fsOrphanScan) Folders() (string, string) {
	return fs.baseFolder, fs.subFolder
}

// SetFolders sets the base folder and subfolder for use with a scanning task
func (fs *fsOrphanScan) SetFolders(baseFolder string, subFolder string) {
	fs.baseFolder = baseFolder
	fs.subFolder = subFolder
}

// Verbose sets the verbosity level of the scanning task
func (fs *fsOrphanScan) Verbose(verbose bool) {
	fs.verbose = verbose
}

// Scan scans for media files which have been removed from the media directory, and removes
// them as appropriate.  An orphan is defined as follows:
//   - Art: art file is no longer present in the filesystem
//   - Artist: no more songs contain this artist's ID
//   - Album: no more songs contain this album's ID
//   - Folder: folder no longer present in the filesystem, or folder contains no items
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

	// Scan for orphans using the specified file source
	return fsSource.OrphanScan(baseFolder, subFolder, fs.verbose, orphanCancelChan)
}
