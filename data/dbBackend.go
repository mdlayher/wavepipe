package data

import (
	"github.com/jmoiron/sqlx"
)

// DB is the current database backend
var DB dbBackend

// dbBackend represents the database backend that the program will connect to
type dbBackend interface {
	Open() (*sqlx.DB, error)
	DSN(string)

	AllArtists() ([]Artist, error)
	PurgeOrphanArtists() (int, error)
	LoadArtist(*Artist) error
	SaveArtist(*Artist) error

	AllAlbums() ([]Album, error)
	AlbumsForArtist(int) ([]Album, error)
	PurgeOrphanAlbums() (int, error)
	DeleteAlbum(*Album) error
	LoadAlbum(*Album) error
	SaveAlbum(*Album) error

	AllSongs() ([]Song, error)
	SongsForAlbum(int) ([]Song, error)
	SongsForArtist(int) ([]Song, error)
	SongsInPath(string) ([]Song, error)
	SongsNotInPath(string) ([]Song, error)
	DeleteSong(*Song) error
	LoadSong(*Song) error
	SaveSong(*Song) error
}
