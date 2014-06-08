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
	db   *sqlx.DB
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

// Open initializes a new sqlite sqlx database connection
func (s *SqliteBackend) Open() error {
	// Open connection using path
	db, err := sqlx.Open("sqlite3", s.Path)
	if err != nil {
		return err
	}

	// Performance tuning

	// Do not wait for OS to respond to data write to disk
	if _, err := db.Exec("PRAGMA synchronous = OFF;"); err != nil {
		return err
	}

	// Keep rollback journal in memory, instead of on disk
	if _, err := db.Exec("PRAGMA journal_mode = MEMORY;"); err != nil {
		return err
	}

	// Store database instance for duration of run
	s.db = db
	return nil
}

// Close closes the current sqlite sqlx database connection
func (s *SqliteBackend) Close() error {
	return s.db.Close()
}

// ArtInPath loads a slice of all Art structs contained within the specified file path
func (s *SqliteBackend) ArtInPath(path string) ([]Art, error) {
	return s.artQuery("SELECT * FROM art WHERE file_name LIKE ?;", path+"%")
}

// ArtNotInPath loads a slice of all Art structs NOT contained within the specified file path
func (s *SqliteBackend) ArtNotInPath(path string) ([]Art, error) {
	return s.artQuery("SELECT * FROM art WHERE file_name NOT LIKE ?;", path+"%")
}

// DeleteArt removes Art from the database
func (s *SqliteBackend) DeleteArt(a *Art) error {
	// Attempt to delete this art by its ID
	tx := s.db.MustBegin()
	tx.Exec("DELETE FROM art WHERE id = ?;", a.ID)

	// Update any songs using this art ID to have a zero ID
	tx.Exec("UPDATE songs SET art_id = 0 WHERE art_id = ?;", a.ID)
	return tx.Commit()
}

// LoadArt loads Art from the database, populating the parameter struct
func (s *SqliteBackend) LoadArt(a *Art) error {
	// Load the artist via ID if available
	if a.ID != 0 {
		if err := s.db.Get(a, "SELECT * FROM art WHERE id = ?;", a.ID); err != nil {
			return err
		}

		return nil
	}

	// Load via file name
	if err := s.db.Get(a, "SELECT * FROM art WHERE file_name = ?;", a.FileName); err != nil {
		return err
	}

	return nil
}

// SaveArt attempts to save Art to the database
func (s *SqliteBackend) SaveArt(a *Art) error {
	// Insert new artist
	query := "INSERT INTO art (`file_name`, `file_size`, `last_modified`) VALUES (?, ?, ?);"
	tx := s.db.MustBegin()
	tx.Exec(query, a.FileName, a.FileSize, a.LastModified)

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	// If no ID, reload to grab it
	if a.ID == 0 {
		if err := s.LoadArt(a); err != nil {
			return err
		}
	}

	return nil
}

// AllArtists loads a slice of all Artist structs from the database
func (s *SqliteBackend) AllArtists() ([]Artist, error) {
	return s.artistQuery("SELECT * FROM artists;")
}

// LimitArtists loads a slice of Artist structs from the database using SQL limit, where the first parameter
// specifies an offset and the second specifies an item count
func (s *SqliteBackend) LimitArtists(offset int, count int) ([]Artist, error) {
	return s.artistQuery("SELECT * FROM artists LIMIT ?, ?;", offset, count)
}

// SearchArtists loads a slice of all Artist structs from the database which contain
// titles that match the specified search query
func (s *SqliteBackend) SearchArtists(query string) ([]Artist, error) {
	return s.artistQuery("SELECT * FROM artists WHERE title LIKE ?;", "%"+query+"%")
}

// PurgeOrphanArtists deletes all artists who are "orphaned", meaning that they no
// longer have any songs which reference their ID
func (s *SqliteBackend) PurgeOrphanArtists() (int, error) {
	// Select all artists without a song referencing their artist ID
	rows, err := s.db.Queryx("SELECT artists.id FROM artists LEFT JOIN songs ON " +
		"artists.id = songs.artist_id WHERE songs.artist_id IS NULL;")
	if err != nil && err != sql.ErrNoRows {
		return -1, err
	}
	defer rows.Close()

	// Open a transaction to remove all orphaned artists
	tx := s.db.MustBegin()

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

	// Error check rows
	if err := rows.Err(); err != nil {
		return -1, err
	}

	return total, tx.Commit()
}

// DeleteArtist removes an Artist from the database
func (s *SqliteBackend) DeleteArtist(a *Artist) error {
	// Attempt to delete this artist by its ID, if available
	tx := s.db.MustBegin()
	if a.ID != 0 {
		tx.Exec("DELETE FROM artists WHERE id = ?;", a.ID)
		return tx.Commit()
	}

	// Else, attempt to remove the artist by its title
	tx.Exec("DELETE FROM artists WHERE title = ?;", a.Title)
	return tx.Commit()
}

// LoadArtist loads an Artist from the database, populating the parameter struct
func (s *SqliteBackend) LoadArtist(a *Artist) error {
	// Load the artist via ID if available
	if a.ID != 0 {
		if err := s.db.Get(a, "SELECT * FROM artists WHERE id = ?;", a.ID); err != nil {
			return err
		}

		return nil
	}

	// Load via title
	if err := s.db.Get(a, "SELECT * FROM artists WHERE title = ?;", a.Title); err != nil {
		return err
	}

	return nil
}

// SaveArtist attempts to save an Artist to the database
func (s *SqliteBackend) SaveArtist(a *Artist) error {
	// Insert new artist
	query := "INSERT INTO artists (`title`) VALUES (?);"
	tx := s.db.MustBegin()
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

// LimitAlbums loads a slice of Album structs from the database using SQL limit, where the first parameter
// specifies an offset and the second specifies an item count
func (s *SqliteBackend) LimitAlbums(offset int, count int) ([]Album, error) {
	return s.albumQuery("SELECT albums.*,artists.title AS artist FROM albums "+
		"JOIN artists ON albums.artist_id = artists.id LIMIT ?, ?;", offset, count)
}

// AlbumsForArtist loads a slice of all Album structs with matching artist ID
func (s *SqliteBackend) AlbumsForArtist(ID int) ([]Album, error) {
	return s.albumQuery("SELECT albums.*,artists.title AS artist FROM albums "+
		"JOIN artists ON albums.artist_id = artists.id WHERE albums.artist_id = ?;", ID)
}

// SearchAlbums loads a slice of all Album structs from the database which contain
// titles that match the specified search query
func (s *SqliteBackend) SearchAlbums(query string) ([]Album, error) {
	return s.albumQuery("SELECT * FROM albums WHERE title LIKE ?;", "%"+query+"%")
}

// PurgeOrphanAlbums deletes all albums who are "orphaned", meaning that they no
// longer have any songs which reference their ID
func (s *SqliteBackend) PurgeOrphanAlbums() (int, error) {
	// Select all albums without a song referencing their album ID
	rows, err := s.db.Queryx("SELECT albums.id FROM albums LEFT JOIN songs ON " +
		"albums.id = songs.album_id WHERE songs.album_id IS NULL;")
	if err != nil && err != sql.ErrNoRows {
		return -1, err
	}
	defer rows.Close()

	// Open a transaction to remove all orphaned albums
	tx := s.db.MustBegin()

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

	// Error check rows
	if err := rows.Err(); err != nil {
		return -1, err
	}

	return total, tx.Commit()
}

// DeleteAlbum removes a Album from the database
func (s *SqliteBackend) DeleteAlbum(a *Album) error {
	// Attempt to delete this album by its ID, if available
	tx := s.db.MustBegin()
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
	// Load the album via ID if available
	if a.ID != 0 {
		if err := s.db.Get(a, "SELECT albums.*,artists.title AS artist FROM albums "+
			"JOIN artists ON albums.artist_id = artists.id WHERE albums.id = ?;", a.ID); err != nil {
			return err
		}

		return nil
	}

	// Load via artist ID and title
	if err := s.db.Get(a, "SELECT albums.*,artists.title AS artist FROM albums "+
		"JOIN artists ON albums.artist_id = artists.id WHERE albums.artist_id = ? AND albums.title = ?;", a.ArtistID, a.Title); err != nil {
		return err
	}

	return nil
}

// SaveAlbum attempts to save an Album to the database
func (s *SqliteBackend) SaveAlbum(a *Album) error {
	// Insert new album
	query := "INSERT INTO albums (`artist_id`, `title`, `year`) VALUES (?, ?, ?);"
	tx := s.db.MustBegin()
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

// AllFolders loads a slice of all Folder structs from the database
func (s *SqliteBackend) AllFolders() ([]Folder, error) {
	return s.folderQuery("SELECT * FROM folders;")
}

// LimitFolders loads a slice of Folder structs from the database using SQL limit, where the first parameter
// specifies an offset and the second specifies an item count
func (s *SqliteBackend) LimitFolders(offset int, count int) ([]Folder, error) {
	return s.folderQuery("SELECT * FROM folders LIMIT ?, ?;", offset, count)
}

// Subfolders loads a slice of all Folder structs residing directly beneath this one from the database
func (s *SqliteBackend) Subfolders(parentID int) ([]Folder, error) {
	return s.folderQuery("SELECT * FROM folders WHERE parent_id = ?;", parentID)
}

// FoldersInPath loads a slice of all Folder structs contained within the specified file path
func (s *SqliteBackend) FoldersInPath(path string) ([]Folder, error) {
	return s.folderQuery("SELECT * FROM folders WHERE path LIKE ?;", path+"%")
}

// FoldersNotInPath loads a slice of all Folder structs NOT contained within the specified file path
func (s *SqliteBackend) FoldersNotInPath(path string) ([]Folder, error) {
	return s.folderQuery("SELECT * FROM folders WHERE path NOT LIKE ?;", path+"%")
}

// SearchFolders loads a slice of all Folder structs from the database which contain
// titles that match the specified search query
func (s *SqliteBackend) SearchFolders(query string) ([]Folder, error) {
	return s.folderQuery("SELECT * FROM folders WHERE title LIKE ?;", "%"+query+"%")
}

// DeleteFolder removes a Folder from the database
func (s *SqliteBackend) DeleteFolder(f *Folder) error {
	// Attempt to delete this folder by its ID, if available
	tx := s.db.MustBegin()
	if f.ID != 0 {
		tx.Exec("DELETE FROM folders WHERE id = ?;", f.ID)
		return tx.Commit()
	}

	// Else, attempt to remove the folder by its path
	tx.Exec("DELETE FROM folders WHERE path = ?;", f.Path)
	return tx.Commit()
}

// LoadFolder loads a Folder from the database, populating the parameter struct
func (s *SqliteBackend) LoadFolder(f *Folder) error {
	// Load the folder via ID if available
	if f.ID != 0 {
		if err := s.db.Get(f, "SELECT * FROM folders WHERE id = ?;", f.ID); err != nil {
			return err
		}

		return nil
	}

	// Load via path
	if err := s.db.Get(f, "SELECT * FROM folders WHERE path = ?;", f.Path); err != nil {
		return err
	}

	return nil
}

// SaveFolder attempts to save an Folder to the database
func (s *SqliteBackend) SaveFolder(f *Folder) error {
	// Insert new folder
	query := "INSERT INTO folders (`parent_id`, `title`, `path`) VALUES (?, ?, ?);"
	tx := s.db.MustBegin()
	tx.Exec(query, f.ParentID, f.Title, f.Path)

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	// If no ID, reload to grab it
	if f.ID == 0 {
		if err := s.LoadFolder(f); err != nil {
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

// LimitSongs loads a slice of Song structs from the database using SQL limit, where the first parameter
// specifies an offset and the second specifies an item count
func (s *SqliteBackend) LimitSongs(offset int, count int) ([]Song, error) {
	return s.songQuery("SELECT songs.*,artists.title AS artist,albums.title AS album FROM songs "+
		"JOIN artists ON songs.artist_id = artists.id JOIN albums ON songs.album_id = albums.id "+
		"LIMIT ?, ?;", offset, count)
}

// RandomSongs loads a slice of 'n' random song structs from the database
func (s *SqliteBackend) RandomSongs(n int) ([]Song, error) {
	return s.songQuery("SELECT songs.*,artists.title AS artist,albums.title AS album FROM songs "+
		"JOIN artists ON songs.artist_id = artists.id JOIN albums ON songs.album_id = albums.id "+
		"ORDER BY RANDOM() LIMIT ?;", n)
}

// SearchSongs loads a slice of all Song structs from the database which contain
// titles that match the specified search query
func (s *SqliteBackend) SearchSongs(query string) ([]Song, error) {
	return s.songQuery("SELECT * FROM songs WHERE title LIKE ?;", "%"+query+"%")
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

// SongsForFolder loads a slice of all Song structs which have the matching folder ID
func (s *SqliteBackend) SongsForFolder(ID int) ([]Song, error) {
	return s.songQuery("SELECT songs.*,artists.title AS artist,albums.title AS album FROM songs "+
		"JOIN artists ON songs.artist_id = artists.id JOIN albums ON songs.album_id = albums.id "+
		"WHERE songs.folder_id = ?;", ID)
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
	// Attempt to delete this song by its ID, if available
	tx := s.db.MustBegin()
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
	// Load the song via ID if available
	if a.ID != 0 {
		if err := s.db.Get(a, "SELECT songs.*,artists.title AS artist,albums.title AS album FROM songs "+
			"JOIN artists ON songs.artist_id = artists.id JOIN albums ON songs.album_id = albums.id "+
			"WHERE songs.id = ?;", a.ID); err != nil {
			return err
		}

		return nil
	}

	// Load via file name
	if err := s.db.Get(a, "SELECT songs.*,artists.title AS artist,albums.title AS album FROM songs "+
		"JOIN artists ON songs.artist_id = artists.id JOIN albums ON songs.album_id = albums.id "+
		"WHERE songs.file_name = ?;", a.FileName); err != nil {
		return err
	}

	return nil
}

// SaveSong attempts to save a Song to the database
func (s *SqliteBackend) SaveSong(a *Song) error {
	// Insert new song
	query := "INSERT INTO songs (`album_id`, `art_id`, `artist_id`, `bitrate`, `channels`, `comment`, `file_name`, " +
		"`file_size`, `file_type_id`, `folder_id`, `genre`, `last_modified`, `length`, `sample_rate`, `title`, `track`, `year`) " +
		" VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);"
	tx := s.db.MustBegin()
	tx.Exec(query, a.AlbumID, a.ArtID, a.ArtistID, a.Bitrate, a.Channels, a.Comment, a.FileName, a.FileSize, a.FileTypeID,
		a.FolderID, a.Genre, a.LastModified, a.Length, a.SampleRate, a.Title, a.Track, a.Year)

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

// UpdateSong attempts to update a Song in the database
func (s *SqliteBackend) UpdateSong(a *Song) error {
	// Update existing song
	query := "UPDATE songs SET `album_id` = ?, `art_id` = ?, `artist_id` = ?, `bitrate` = ?, `channels` = ?, `comment` = ?, " +
		"`file_size` = ?, `folder_id` = ?,  `genre` = ?, `last_modified` = ?, `length` = ?, `sample_rate` = ?, " +
		"`title` = ?, `track` = ?, `year` = ? WHERE `id` = ?;"
	tx := s.db.MustBegin()
	tx.Exec(query, a.AlbumID, a.ArtID, a.ArtistID, a.Bitrate, a.Channels, a.Comment, a.FileSize,
		a.FolderID, a.Genre, a.LastModified, a.Length, a.SampleRate, a.Title, a.Track, a.Year, a.ID)

	// Commit transaction
	return tx.Commit()
}

// DeleteUser removes a User from the database
func (s *SqliteBackend) DeleteUser(u *User) error {
	// Attempt to delete this user by its ID, if available
	tx := s.db.MustBegin()
	if u.ID != 0 {
		tx.Exec("DELETE FROM users WHERE id = ?;", u.ID)
		return tx.Commit()
	}

	// Else, attempt to remove the user by its username
	tx.Exec("DELETE FROM users WHERE username = ?;", u.Username)
	return tx.Commit()
}

// LoadUser loads a User from the database, populating the parameter struct
func (s *SqliteBackend) LoadUser(u *User) error {
	// Load the user via ID if available
	if u.ID != 0 {
		if err := s.db.Get(u, "SELECT * FROM users WHERE id = ?;", u.ID); err != nil {
			return err
		}

		return nil
	}

	// Load via username
	if err := s.db.Get(u, "SELECT * FROM users WHERE username = ?;", u.Username); err != nil {
		return err
	}

	return nil
}

// SaveUser attempts to save a User to the database
func (s *SqliteBackend) SaveUser(u *User) error {
	// Insert new user
	query := "INSERT INTO users (`username`, `password`, `lastfm_token`) VALUES (?, ?, ?);"
	tx := s.db.MustBegin()
	tx.Exec(query, u.Username, u.Password, u.LastFMToken)

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	// If no ID, reload to grab it
	if u.ID == 0 {
		if err := s.LoadUser(u); err != nil {
			return err
		}
	}

	return nil
}

// UpdateUser updates a User in the database
func (s *SqliteBackend) UpdateUser(u *User) error {
	// Attempt to update this user by its ID, if available
	tx := s.db.MustBegin()
	if u.ID != 0 {
		tx.Exec("UPDATE users SET `username` = ?, `password` = ?, `lastfm_token` = ? WHERE id = ?;",
			u.Username, u.Password, u.LastFMToken, u.ID)
		return tx.Commit()
	}

	// Else, attempt to update the user by its username
	tx.Exec("UPDATE users SET `password` = ?, `lastfm_token` = ? WHERE username = ?;",
		u.Password, u.LastFMToken, u.Username)
	return tx.Commit()
}

// DeleteSession removes a Session from the database
func (s *SqliteBackend) DeleteSession(u *Session) error {
	// Attempt to delete this session by its ID, if available
	tx := s.db.MustBegin()
	if u.ID != 0 {
		tx.Exec("DELETE FROM sessions WHERE id = ?;", u.ID)
		return tx.Commit()
	}

	// Else, attempt to remove the session by its key
	tx.Exec("DELETE FROM sessions WHERE key = ?;", u.Key)
	return tx.Commit()
}

// LoadSession loads a Session from the database, populating the parameter struct
func (s *SqliteBackend) LoadSession(u *Session) error {
	// Load the session via ID if available
	if u.ID != 0 {
		if err := s.db.Get(u, "SELECT * FROM sessions WHERE id = ?;", u.ID); err != nil {
			return err
		}

		return nil
	}

	// Load via key
	if err := s.db.Get(u, "SELECT * FROM sessions WHERE key = ?;", u.Key); err != nil {
		return err
	}

	return nil
}

// SaveSession attempts to save a Session to the database
func (s *SqliteBackend) SaveSession(u *Session) error {
	// Insert new session
	query := "INSERT INTO sessions (`user_id`, `client`, `expire`, `key`) VALUES (?, ?, ?, ?);"
	tx := s.db.MustBegin()
	tx.Exec(query, u.UserID, u.Client, u.Expire, u.Key)

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	// If no ID, reload to grab it
	if u.ID == 0 {
		if err := s.LoadSession(u); err != nil {
			return err
		}
	}

	return nil
}

// UpdateSession updates a Session in the database
func (s *SqliteBackend) UpdateSession(u *Session) error {
	// Attempt to update this session by its ID, if available
	tx := s.db.MustBegin()
	if u.ID != 0 {
		tx.Exec("UPDATE sessions SET `expire` = ? WHERE id = ?;", u.Expire, u.ID)
		return tx.Commit()
	}

	// Else, attempt to update the session by its key
	tx.Exec("UPDATE sessions SET `expire` = ? WHERE key = ?;", u.Expire, u.Key)
	return tx.Commit()
}

// albumQuery loads a slice of Album structs matching the input query
func (s *SqliteBackend) albumQuery(query string, args ...interface{}) ([]Album, error) {
	// Perform input query with arguments
	rows, err := s.db.Queryx(query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

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

	// Error check rows
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return albums, nil
}

// artQuery loads a slice of Art structs matching the input query
func (s *SqliteBackend) artQuery(query string, args ...interface{}) ([]Art, error) {
	// Perform input query with arguments
	rows, err := s.db.Queryx(query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	// Iterate all rows
	art := make([]Art, 0)
	a := Art{}
	for rows.Next() {
		// Scan artist into struct
		if err := rows.StructScan(&a); err != nil {
			return nil, err
		}

		// Append to list
		art = append(art, a)
	}

	// Error check rows
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return art, nil
}

// artistQuery loads a slice of Artist structs matching the input query
func (s *SqliteBackend) artistQuery(query string, args ...interface{}) ([]Artist, error) {
	// Perform input query with arguments
	rows, err := s.db.Queryx(query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

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

	// Error check rows
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return artists, nil
}

// folderQuery loads a slice of Folder structs matching the input query
func (s *SqliteBackend) folderQuery(query string, args ...interface{}) ([]Folder, error) {
	// Perform input query with arguments
	rows, err := s.db.Queryx(query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

	// Iterate all rows
	folders := make([]Folder, 0)
	a := Folder{}
	for rows.Next() {
		// Scan folder into struct
		if err := rows.StructScan(&a); err != nil {
			return nil, err
		}

		// Append to list
		folders = append(folders, a)
	}

	// Error check rows
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return folders, nil
}

// songQuery loads a slice of Song structs matching the input query
func (s *SqliteBackend) songQuery(query string, args ...interface{}) ([]Song, error) {
	// Perform input query with arguments
	rows, err := s.db.Queryx(query, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()

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

	// Error check rows
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return songs, nil
}
