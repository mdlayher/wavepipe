package core

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// sqliteBackend represents a sqlite3-based database backend
type sqliteBackend struct {
	Path string
}

// DSN sets the Path for use with sqlite3
func (s *sqliteBackend) DSN(path string) {
	s.Path = path
}

// Open opens a new sqlx database connection
func (s *sqliteBackend) Open() (*sqlx.DB, error) {
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

// LoadArtist loads an Artist from the database, populating the parameter struct
func (s *sqliteBackend) LoadArtist(a *Artist) error {
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
func (s *sqliteBackend) SaveArtist(a *Artist) error {
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

// LoadAlbum loads an Album from the database, populating the parameter struct
func (s *sqliteBackend) LoadAlbum(a *Album) error {
	// Open database
	db, err := s.Open()
	if err != nil {
		return err
	}
	defer db.Close()

	// Load the album via ID if available
	if a.ID != 0 {
		if err := db.Get(a, "SELECT * FROM albums WHERE id = ?;", a.ID); err != nil {
			return err
		}

		return nil
	}

	// Load via artist ID and title
	if err := db.Get(a, "SELECT * FROM albums WHERE artist_id = ? AND title = ?;", a.ArtistID, a.Title); err != nil {
		return err
	}

	return nil
}

// SaveAlbum attempts to save an Album to the database
func (s *sqliteBackend) SaveAlbum(a *Album) error {
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

// LoadSong loads an Song from the database, populating the parameter struct
func (s *sqliteBackend) LoadSong(a *Song) error {
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

	// Load via file name and title
	if err := db.Get(a, "SELECT * FROM songs WHERE file_name = ? AND title = ?;", a.FileName, a.Title); err != nil {
		return err
	}

	return nil
}

// SaveSong attempts to save an Song to the database
func (s *sqliteBackend) SaveSong(a *Song) error {
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
