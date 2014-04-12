package core

import (
	"log"

	"github.com/jmoiron/sqlx"

	// Include sqlite3 driver
	_ "github.com/mattn/go-sqlite3"
)

// dbBackend represents the database backend that the program will connect to
type dbBackend interface {
	Open() (*sqlx.DB, error)
	DSN(string)
	AllArtists() ([]Artist, error)
	PurgeOrphanArtists() (int, error)
	LoadArtist(*Artist) error
	SaveArtist(*Artist) error
	AllAlbums() ([]Album, error)
	PurgeOrphanAlbums() (int, error)
	LoadAlbum(*Album) error
	SaveAlbum(*Album) error
	AllSongs() ([]Song, error)
	SongsInPath(string) ([]Song, error)
	DeleteSong(*Song) error
	LoadSong(*Song) error
	SaveSong(*Song) error
}

// db is the current database backend
var db dbBackend

// dbManager manages database connections, and communicates back and forth with the manager goroutine
func dbManager(dbPath string, dbKillChan chan struct{}) {
	log.Println("db: starting...")

	// Attempt to open database connection
	if dbPath != "" {
		db = new(sqliteBackend)
		db.DSN(dbPath)
	}

	// Trigger events via channel
	for {
		select {
		// Stop database manager
		case <-dbKillChan:
			// Inform manager that shutdown is complete
			log.Println("db: stopped!")
			dbKillChan <- struct{}{}
			return
		}
	}
}
