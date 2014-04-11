package core

import (
	"database/sql"

	"github.com/jmoiron/sqlx"

	// Include sqlite3 driver
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

// AllArtists loads a slice of all Artist structs from the database
func (s *sqliteBackend) AllArtists() ([]*Artist, error) {
	// Open database
	db, err := s.Open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Query for a list of all artists
	rows, err := db.Queryx("SELECT * FROM artists;")
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Iterate all rows
	artists := make([]*Artist, 0)
	a := new(Artist)
	for rows.Next() {
		// Scan artist into struct
		if err := rows.StructScan(a); err != nil {
			return nil, err
		}

		// Append to list
		artists = append(artists, a)
	}

	return artists, nil
}

// PurgeOrphanArtists deletes all artists who are "orphaned", meaning that they no
// longer have any songs which reference their ID
func (s *sqliteBackend) PurgeOrphanArtists() (int, error) {
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
