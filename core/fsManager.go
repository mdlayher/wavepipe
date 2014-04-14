package core

import (
	"errors"
	"log"
	"os"

	"github.com/mdlayher/goset"
	"github.com/romanoff/fsmonitor"
)

// validSet is a set of valid file extensions which we should scan as media, as they are the ones
// which TagLib is capable of reading
var validSet = set.New(".ape", ".flac", ".m4a", ".mp3", ".mpc", ".ogg", ".wma", ".wv")

// fsQueue is a queue of tasks to be performed by the filesystem, such as media and orphan scans
var fsQueue = make(chan fsTask, 10)

// fsSource is the data source used to scan for media files (could be filesystem, memory, etc)
var fsSource fileSource

// fsTask is the interface which defines a filesystem task, such as a media scan, or an orphan scan
type fsTask interface {
	Folders() (string, string)
	SetFolders(string, string)
	Scan(string, string, chan struct{}) error
}

// fileSource represents a source from which files can be scanned and indexed
type fileSource interface {
	MediaScan(string, chan struct{}) error
	OrphanScan(string, string, chan struct{}) error
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

	// Queue an orphan scan
	o := new(fsOrphanScan)
	o.SetFolders(mediaFolder, "")
	fsQueue <- o

	// Queue a media scan
	m := new(fsMediaScan)
	m.SetFolders(mediaFolder, "")
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
			}
		}
	}()

	// Create a filesystem watcher
	watcher, err := fsmonitor.NewWatcher()
	if err != nil {
		log.Println(err)
		return
	}

	// Wait for events on goroutine
	// TODO: for the time being, we ignore IsModify() and IsRename() events, because the
	// IsCreate() and IsDelete() methods, and recurring scans will cover these.  Will re-evalaute
	// this later.
	go func() {
		for {
			select {
			// Event occurred
			case ev := <-watcher.Event:
				switch {
				// On create, trigger a media scan
				case ev.IsCreate():
					// Scan item as the "base folder", so it just adds this item
					m := new(fsMediaScan)
					m.SetFolders(ev.Name, "")
					fsQueue <- m
				// On delete, trigger an orphan scan
				case ev.IsDelete():
					// Scan item as the "subfolder", so it just removes this item
					o := new(fsOrphanScan)
					o.SetFolders("", ev.Name)
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

// Scan scans for media files in a specified path, and queues them up for inclusion
// in the wavepipe database
func (fs *fsMediaScan) Scan(mediaFolder string, subFolder string, walkCancelChan chan struct{}) error {
	// Media scans are comprehensive, so subfolder has no purpose
	if subFolder != "" {
		return errors.New("media scan: subfolder not valid for media scan operation")
	}

	// Scan for media using the specified file source
	return fsSource.MediaScan(mediaFolder, walkCancelChan)
}

// fsOrphanScan represents a filesystem task which scans the given path for orphaned media
type fsOrphanScan struct {
	baseFolder string
	subFolder  string
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

	// Scan for orphans using the specified file source
	return fsSource.OrphanScan(baseFolder, subFolder, orphanCancelChan)
}
