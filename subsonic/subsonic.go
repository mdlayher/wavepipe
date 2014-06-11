package subsonic

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"time"

	"github.com/mdlayher/wavepipe/api"
	"github.com/mdlayher/wavepipe/data"

	"github.com/gorilla/context"
	"github.com/unrolled/render"
)

const (
	// XMLName is the top-level name of a Subsonic XML response
	XMLName = "subsonic-response"
	// XMLNS is the XML namespace of a Subsonic XML response
	XMLNS = "http://subsonic.org/restapi"
	// Version is the emulated Subsonic API version
	Version = "1.8.0"
)

var (
	// ErrBadCredentials returns a bad credentials response
	ErrBadCredentials = func() *Container {
		// Generate new container with failed status
		c := newContainer()
		c.Status = "failed"

		// Return error
		c.SubError = &Error{Code: 40, Message: "Wrong username or password."}
		return c
	}()
	// ErrGeneric returns a generic error response, such as a server issue
	ErrGeneric = func() *Container {
		// Generate new container with failed status
		c := newContainer()
		c.Status = "failed"

		// Return error
		c.SubError = &Error{Code: 0, Message: "An error occurred."}
		return c
	}()
	// ErrMissingParameter returns a missing required parameter response
	ErrMissingParameter = func() *Container {
		// Generate new container with failed status
		c := newContainer()
		c.Status = "failed"

		// Return error
		c.SubError = &Error{Code: 10, Message: "Required parameter is missing."}
		return c
	}()
)

// newContainer creates a new, empty Container with the proper attributes
func newContainer() *Container {
	return &Container{
		XMLNS:   XMLNS,
		Status:  "ok",
		Version: Version,
	}
}

// Container is the top-level emulated Subsonic response
type Container struct {
	// Top-level container name
	XMLName xml.Name `xml:"subsonic-response"`

	// Attributes which are always present
	XMLNS   string `xml:"xmlns,attr"`
	Status  string `xml:"status,attr"`
	Version string `xml:"version,attr"`

	// Error, returned on failures
	SubError *Error

	// Nested data

	// getAlbum.view
	Album []Album `xml:"album"`

	// getAlbumList2.view
	AlbumList2 *AlbumList2Container

	// getMusicFolders.view
	MusicFolders *MusicFoldersContainer

	// getRandomSongs.view
	RandomSongs *RandomSongsContainer
}

// Error returns the error code and message from Subsonic, and enables Subsonic
// errors to be returned in authentication
func (c Container) Error() string {
	return fmt.Sprintf("%d: %s", c.SubError.Code, c.SubError.Message)
}

// Error contains a Subsonic error, with status code and message
type Error struct {
	XMLName xml.Name `xml:"error,omitempty"`

	Code    int    `xml:"code,attr"`
	Message string `xml:"message,attr"`
}

// Album represents an emulated Subsonic album
type Album struct {
	// Subsonic fields
	ID        int    `xml:"id,attr"`
	Name      string `xml:"name,attr"`
	Artist    string `xml:"artist,attr"`
	ArtistID  int    `xml:"artistId,attr"`
	CoverArt  string `xml:"coverArt,attr"`
	SongCount int    `xml:"songCount,attr"`
	Duration  int    `xml:"duration,attr"`
	Created   string `xml:"created,attr"`

	// Nested data

	// getAlbum.view
	Songs []Song `xml:"song"`
}

// subAlbum turns a wavepipe album and songs into a Subsonic format album
func subAlbum(album data.Album, songs data.SongSlice) Album {
	return Album{
		ID:        album.ID,
		Name:      album.Title,
		Artist:    album.Artist,
		ArtistID:  album.ArtistID,
		CoverArt:  strconv.Itoa(songs[0].ArtID),
		SongCount: len(songs),
		Duration:  songs.Length(),
		Created:   time.Unix(songs[0].LastModified, 0).Format("2006-01-02T15:04:05"),
	}
}

// Song represents an emulated Subsonic song
type Song struct {
	ID          int    `xml:"id,attr"`
	Parent      int    `xml:"parent,attr"`
	Title       string `xml:"title,attr"`
	Album       string `xml:"album,attr"`
	Artist      string `xml:"artist,attr"`
	IsDir       bool   `xml:"isDir,attr"`
	CoverArt    string `xml:"coverArt,attr"`
	Created     string `xml:"created,attr"`
	Duration    int    `xml:"duration,attr"`
	BitRate     int    `xml:"bitRate,attr"`
	Track       int    `xml:"track,attr"`
	DiscNumber  int    `xml:"discNumber,attr"`
	Year        int    `xml:"year,attr"`
	Genre       string `xml:"genre,attr"`
	Size        int64  `xml:"size,attr"`
	Suffix      string `xml:"suffix,attr"`
	ContentType string `xml:"contentType,attr"`
	IsVideo     bool   `xml:"isVideo,attr"`
	Path        string `xml:"path,attr"`
	AlbumID     int    `xml:"albumId,attr"`
	ArtistID    int    `xml:"artistId,attr"`
	Type        string `xml:"type,attr"`
}

// subSong turns a wavepipe song into a Subsonic format song
func subSong(song data.Song) Song {
	return Song{
		ID: song.ID,
		// BUG(mdlayher): subsonic: wavepipe has no concept of a parent item, so leave blank?
		Parent:   0,
		Title:    song.Title,
		Album:    song.Album,
		Artist:   song.Artist,
		IsDir:    false,
		CoverArt: strconv.Itoa(song.ArtID),
		Created:  time.Unix(song.LastModified, 0).Format("2006-01-02T15:04:05"),
		Duration: song.Length,
		BitRate:  song.Bitrate,
		Track:    song.Track,
		// BUG(mdlayher): subsonic: wavepipe cannot scan disc number without taggolib
		DiscNumber:  1,
		Year:        song.Year,
		Genre:       song.Genre,
		Size:        song.FileSize,
		Suffix:      path.Ext(song.FileName)[1:],
		ContentType: data.MIMEMap[song.FileTypeID],
		IsVideo:     false,
		Path:        song.FileName,
		AlbumID:     song.AlbumID,
		ArtistID:    song.ArtistID,
		Type:        "music",
	}
}

// GetPing is used in Subsonic to check server connectivity
func GetPing(res http.ResponseWriter, req *http.Request) {
	// Retrieve render
	r := context.Get(req, api.CtxRender).(*render.Render)

	// Output blank container
	r.XML(res, 200, newContainer())
}

// MusicFolder represents an emulated Subsonic music folder
type MusicFolder struct {
	ID   int    `xml:"id,attr"`
	Name string `xml:"name,attr"`
}
