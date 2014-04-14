package data

import (
	"database/sql"
	"errors"
	"io"
	"log"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/jmoiron/sqlx"

	// Include sqlite3 driver
	_ "github.com/mattn/go-sqlite3"
)

// SqliteBackend represents a sqlite3-based database backend
type SqliteBackend struct {
	Path string
}

// DSN sets the Path for use with sqlite3
func (s *SqliteBackend) DSN(path string) {
	// Get current user
	user, err := user.Current()
	if err != nil {
		log.Println(err)
		s.Path = path
		return
	}

	// Replace the home character to set path
	s.Path = strings.Replace(path, "~", user.HomeDir, -1)
}

// Setup copies the empty sqlite database into the wavepipe configuration directory
func (s *SqliteBackend) Setup() error {
	// Check for configuration at this path
	_, err := os.Stat(s.Path)
	if err == nil {
		// Database file exists
		return nil
	}

	// If error is something other than file not exists, return
	if !os.IsNotExist(err) {
		return err
	}

	// Get current user
	user, err := user.Current()
	if err != nil {
		return err
	}

	// Only create file if it's in the default location
	if s.Path != user.HomeDir+"/.config/wavepipe/wavepipe.db" {
		return errors.New("db: cannot create database file: " + s.Path)
	}

	log.Println("db: creating new database file:", s.Path)

	// Create a new config file in the default location
	dir := path.Dir(s.Path) + "/"
	file := path.Base(s.Path)

	// Make directory
	if err := os.MkdirAll(dir, 0775); err != nil {
		return err
	}

	// Grab GOPATH, use only the first path
	gopath := strings.Split(os.Getenv("GOPATH"), ":")[0]

	// Attempt to open database
	src, err := os.Open(gopath + "/src/github.com/mdlayher/wavepipe/res/sqlite/" + file)
	if err != nil {
		return err
	}

	// Attempt to open destination
	dest, err := os.Create(dir + file)
	if err != nil {
		return err
	}

	// Copy contents into destination
	if _, err := io.Copy(dest, src); err != nil {
		return err
	}

	// Close files
	if err := src.Close(); err != nil {
		return err
	}
	if err := dest.Close(); err != nil {
		return err
	}

	return nil
}

// Open opens a new sqlx database connection
func (s *SqliteBackend) Open() (*sqlx.DB, error) {
	// Open connection using path
	db, err := sqlx.Open("sqlite3", s.Path)
	if err != nil {
		return nil, err
	}

	// Performance tuning
	if _, err := db.Exec("PRAGMA synchronous = OFF;"); err != nil {
		return nil, err
	}

	return db, nil
}

// AllArtists loads a slice of all Artist structs from the database
func (s *SqliteBackend) AllArtists() ([]Artist, error) {
	return s.artistQuery("SELECT * FROM artists;")
}

// PurgeOrphanArtists deletes all artists who are "orphaned", meaning that they no
// longer have any songs which reference their ID
func (s *SqliteBackend) PurgeOrphanArtists() (int, error) {
	// Open database
	db, err := s.Open()
	if err != nil {
		return -1, err
	}
	defer db.Close()

	// Select all artists without a song referencing their artist ID
	rows, err := db.Queryx("SELECT artists.id FROM artists LEFT JOIN songs ON " +
		"artists.id = songs.artist_id WHERE songs.artist_id IS NULL;")
	if err != nil && err != sql.ErrNoRows {
		return -1, err
	}

	// Open a transaction to remove all orphaned artists
	tx := db.MustBegin()

	// Iterate all rows
	artist := new(Artist)
	total := 0
	for rows.Next() {
		// Scan ID into struct
		if err := rows.StructScan(artist); err != nil {
			return -1, err
		}

		// Remove artist
		tx.Exec("DELETE FROM artists WHERE id = ?;", artist.ID)
		total++
	}

	return total, tx.Commit()
}

// LoadArtist loads an Artist from the database, populating the parameter struct
func (s *SqliteBackend) LoadArtist(a *Artist) error {
	// Open database
	db, err := s.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	// Load the artist via ID if available
	if a.ID != 0 {
		if err := db.Get(a, "SELECT * FROM artists WHERE id = ?;", a.ID); err != nil {
			return err
		}

		return nil
	}

	// Load via title
	if err := db.Get(a, "SELECT * FROM artists WHERE title = ?;", a.Title); err != nil {
		return err
	}

	return nil
}

// SaveArtist attempts to save an Artist to the database
func (s *SqliteBackend) SaveArtist(a *Artist) error {
	// Open database
	db, err := s.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	// Insert new artist
	query := "INSERT INTO artists (`title`) VALUES (?);"
	tx := db.MustBegin()
	tx.Exec(query, a.Title)

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	// If no ID, reload to grab it
	if a.ID == 0 {
		if err := s.LoadArtist(a); err != nil {
			return err
		}
	}

	return nil
}

// AllAlbums loads a slice of all Album structs from the database
func (s *SqliteBackend) AllAlbums() ([]Album, error) {
	return s.albumQuery("SELECT albums.*,artists.title AS artist FROM albums " +
		"JOIN artists ON albums.artist_id = artists.id;")
}

// AlbumsForArtist loads a slice of all Album structs with matching artist ID
func (s *SqliteBackend) AlbumsForArtist(ID int) ([]Album, error) {
	return s.albumQuery("SELECT albums.*,artists.title AS artist FROM albums "+
		"JOIN artists ON albums.artist_id = artists.id WHERE albums.artist_id = ?;", ID)
}

// PurgeOrphanAlbums deletes all albums who are "orphaned", meaning that they no
// longer have any songs which reference their ID
func (s *SqliteBackend) PurgeOrphanAlbums() (int, error) {
	// Open database
	db, err := s.Open()
	if err != nil {
		return -1, err
	}
	defer db.Close()

	// Select all albums without a song referencing their album ID
	rows, err := db.Queryx("SELECT albums.id FROM albums LEFT JOIN songs ON " +
		"albums.id = songs.album_id WHERE songs.album_id IS NULL;")
	if err != nil && err != sql.ErrNoRows {
		return -1, err
	}

	// Open a transaction to remove all orphaned albums
	tx := db.MustBegin()

	// Iterate all rows
	album := new(Album)
	total := 0
	for rows.Next() {
		// Scan ID into struct
		if err := rows.StructScan(album); err != nil {
			return -1, err
		}

		// Remove album
		tx.Exec("DELETE FROM albums WHERE id = ?;", album.ID)
		total++
	}

	return total, tx.Commit()
}

// DeleteAlbum removes a Album from the database
func (s *SqliteBackend) DeleteAlbum(a *Album) error {
	// Open database
	db, err := s.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	// Attempt to delete this album by its ID, if available
	tx := db.MustBegin()
	if a.ID != 0 {
		tx.Exec("DELETE FROM albums WHERE id = ?;", a.ID)
		return tx.Commit()
	}

	// Else, attempt to remove the album by its artist ID and title
	tx.Exec("DELETE FROM albums WHERE artist_id = ? AND title = ?;", a.ArtistID, a.Title)
	return tx.Commit()
}

// LoadAlbum loads an Album from the database, populating the parameter struct
func (s *SqliteBackend) LoadAlbum(a *Album) error {
	// Open database
	db, err := s.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	// Load the album via ID if available
	if a.ID != 0 {
		if err := db.Get(a, "SELECT albums.*,artists.title AS artist FROM albums "+
			"JOIN artists ON albums.artist_id = artists.id WHERE albums.id = ?;", a.ID); err != nil {
			return err
		}

		return nil
	}

	// Load via artist ID and title
	if err := db.Get(a, "SELECT albums.*,artists.title AS artist FROM albums "+
		"JOIN artists ON albums.artist_id = artists.id WHERE albums.artist_id = ? AND albums.title = ?;", a.ArtistID, a.Title); err != nil {
		return err
	}

	return nil
}

// SaveAlbum attempts to save an Album to the database
func (s *SqliteBackend) SaveAlbum(a *Album) error {
	// Open database
	db, err := s.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	// Insert new album
	query := "INSERT INTO albums (`artist_id`, `title`, `year`) VALUES (?, ?, ?);"
	tx := db.MustBegin()
	tx.Exec(query, a.ArtistID, a.Title, a.Year)

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	// If no ID, reload to grab it
	if a.ID == 0 {
		if err := s.LoadAlbum(a); err != nil {
			return err
		}
	}

	return nil
}

// AllSongs loads a slice of all Song structs from the database
func (s *SqliteBackend) AllSongs() ([]Song, error) {
	return s.songQuery("SELECT songs.*,artists.title AS artist,albums.title AS album FROM songs " +
		"JOIN artists ON songs.artist_id = artists.id JOIN albums ON songs.album_id = albums.id;")
}

// SongsForAlbum loads a slice of all Song structs which have the matching album ID
func (s *SqliteBackend) SongsForAlbum(ID int) ([]Song, error) {
	return s.songQuery("SELECT songs.*,artists.title AS artist,albums.title AS album FROM songs "+
		"JOIN artists ON songs.artist_id = artists.id JOIN albums ON songs.album_id = albums.id "+
		"WHERE songs.album_id = ?;", ID)
}

// SongsForArtist loads a slice of all Song structs which have the matching artist ID
func (s *SqliteBackend) SongsForArtist(ID int) ([]Song, error) {
	return s.songQuery("SELECT songs.*,artists.title AS artist,albums.title AS album FROM songs "+
		"JOIN artists ON songs.artist_id = artists.id JOIN albums ON songs.album_id = albums.id "+
		"WHERE songs.artist_id = ?;", ID)
}

// SongsInPath loads a slice of all Song structs residing under the specified
// filesystem path from the database
func (s *SqliteBackend) SongsInPath(path string) ([]Song, error) {
	return s.songQuery("SELECT songs.*,artists.title AS artist,albums.title AS album FROM songs "+
		"JOIN artists ON songs.artist_id = artists.id JOIN albums ON songs.album_id = albums.id "+
		"WHERE songs.file_name LIKE ?;", path+"%")
}

// SongsNotInPath loads a slice of all Song structs that do not reside under the specified
// filesystem path from the database
func (s *SqliteBackend) SongsNotInPath(path string) ([]Song, error) {
	return s.songQuery("SELECT songs.*,artists.title AS artist,albums.title AS album FROM songs "+
		"JOIN artists ON songs.artist_id = artists.id JOIN albums ON songs.album_id = albums.id "+
		"WHERE songs.file_name NOT LIKE ?;", path+"%")
}

// DeleteSong removes a Song from the database
func (s *SqliteBackend) DeleteSong(a *Song) error {
	// Open database
	db, err := s.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	// Attempt to delete this song by its ID, if available
	tx := db.MustBegin()
	if a.ID != 0 {
		tx.Exec("DELETE FROM songs WHERE id = ?;", a.ID)
		return tx.Commit()
	}

	// Else, attempt to remove the song by its file name
	tx.Exec("DELETE FROM songs WHERE file_name = ?;", a.FileName)
	return tx.Commit()
}

// LoadSong loads a Song from the database, populating the parameter struct
func (s *SqliteBackend) LoadSong(a *Song) error {
	// Open database
	db, err := s.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	// Load the song via ID if available
	if a.ID != 0 {
		if err := db.Get(a, "SELECT * FROM songs WHERE id = ?;", a.ID); err != nil {
			return err
		}

		return nil
	}

	// Load via file name
	if err := db.Get(a, "SELECT * FROM songs WHERE file_name = ?;", a.FileName); err != nil {
		return err
	}

	return nil
}

// SaveSong attempts to save a Song to the database
func (s *SqliteBackend) SaveSong(a *Song) error {
	// Open database
	db, err := s.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	// Insert new song
	query := "INSERT INTO songs (`album_id`, `artist_id`, `bitrate`, `channels`, `comment`, `file_name`, " +
		"`file_size`, `file_type`, `genre`, `last_modified`, `length`, `sample_rate`, `title`, `track`, `year`) " +
		" VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);"
	tx := db.MustBegin()
	tx.Exec(query, a.AlbumID, a.ArtistID, a.Bitrate, a.Channels, a.Comment, a.FileName, a.FileSize, a.FileType,
		a.Genre, a.LastModified, a.Length, a.SampleRate, a.Title, a.Track, a.Year)

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	// If no ID, reload to grab it
	if a.ID == 0 {
		if err := s.LoadSong(a); err != nil {
			return err
		}
	}

	return nil
}

// albumQuery loads a slice of Album structs matching the input query
func (s *SqliteBackend) albumQuery(query string, args ...interface{}) ([]Album, error) {
	// Open database
	db, err := s.Open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Perform input query with arguments
	rows, err := db.Queryx(query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Iterate all rows
	albums := make([]Album, 0)
	a := Album{}
	for rows.Next() {
		// Scan album into struct
		if err := rows.StructScan(&a); err != nil {
			return nil, err
		}

		// Append to list
		albums = append(albums, a)
	}

	return albums, nil
}

// artistQuery loads a slice of Artist structs matching the input query
func (s *SqliteBackend) artistQuery(query string, args ...interface{}) ([]Artist, error) {
	// Open database
	db, err := s.Open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Perform input query with arguments
	rows, err := db.Queryx(query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Iterate all rows
	artists := make([]Artist, 0)
	a := Artist{}
	for rows.Next() {
		// Scan artist into struct
		if err := rows.StructScan(&a); err != nil {
			return nil, err
		}

		// Append to list
		artists = append(artists, a)
	}

	return artists, nil
}

// songQuery loads a slice of Song structs matching the input query
func (s *SqliteBackend) songQuery(query string, args ...interface{}) ([]Song, error) {
	// Open database
	db, err := s.Open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Perform input query with arguments
	rows, err := db.Queryx(query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Iterate all rows
	songs := make([]Song, 0)
	a := Song{}
	for rows.Next() {
		// Scan song into struct
		if err := rows.StructScan(&a); err != nil {
			return nil, err
		}

		// Append to list
		songs = append(songs, a)
	}

	return songs, nil
}
