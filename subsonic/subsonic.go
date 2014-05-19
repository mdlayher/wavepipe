package subsonic

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mdlayher/wavepipe/api"
	"github.com/mdlayher/wavepipe/data"

	"github.com/martini-contrib/render"
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

// AlbumList2Container contains a list of emulated Subsonic albums, by tags
type AlbumList2Container struct {
	// Container name
	XMLName xml.Name `xml:"albumList2,omitempty"`

	// Albums
	Albums []Album `xml:"album"`
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

// newContainer creates a new, empty Container with the proper attributes
func newContainer() *Container {
	return &Container{
		XMLNS:   XMLNS,
		Status:  "ok",
		Version: Version,
	}
}

// GetPing is used in Subsonic to check server connectivity
func GetPing(res http.ResponseWriter) {
	// All Subsonic emulation replies are XML
	res.Header().Set("Content-Type", "text/xml")

	// Marshal empty container to XML
	out, err := xml.Marshal(newContainer())
	if err != nil {
		return
	}

	// Remove closing tag, replace with self-closing tag (needed by Android client)
	res.Write(bytes.Replace(out, []byte("></"+XMLName+">"), []byte("/>"), -1))
}

// GetAlbumList2 is used in Subsonic to return a list of albums organized with tags
func GetAlbumList2(req *http.Request, res http.ResponseWriter, r render.Render) {
	// Create a new response container
	c := newContainer()

	// Fetch all albums
	// TODO: add a LimitAlbums method to fetch subsets
	albums, err := data.DB.AllAlbums()
	if err != nil {
		log.Println(err)
		r.XML(200, ErrGeneric)
		return
	}

	// If offset is past albums count, stop sending albums
	qOffset := req.URL.Query().Get("offset")
	if qOffset != "" {
		// Parse offset
		offset, err := strconv.Atoi(qOffset)
		if err != nil {
			log.Println(err)
			r.XML(200, ErrGeneric)
			return
		}

		// Check if offset is greater than count
		if offset > len(albums) {
			// Empty albums list
			c.AlbumList2 = new(AlbumList2Container)

			// Write empty response
			r.XML(200, c)
			return
		}
	}

	// Iterate all albums
	outAlbums := make([]Album, 0)
	for _, a := range albums {
		// Load songs for album
		songs, err := data.DB.SongsForAlbum(a.ID)
		if err != nil {
			log.Println(err)
			r.XML(200, ErrGeneric)
			return
		}

		// Get cover art, duration, and creation time from songs
		coverArt := 0
		duration := 0
		created := int64(0)

		// Sum up duration
		for i, s := range songs {
			duration += s.Length

			// Set cover art and created time from first song
			if i == 0 {
				coverArt = s.ArtID
				created = s.LastModified
			}
		}

		// Append Subsonic-style album to list
		outAlbums = append(outAlbums, Album{
			ID:        a.ID,
			Name:      a.Title,
			Artist:    a.Artist,
			ArtistID:  a.ArtistID,
			CoverArt:  strconv.Itoa(coverArt),
			SongCount: len(songs),
			Duration:  duration,
			Created:   time.Unix(created, 0).Format("2006-01-02T15:04:05"),
		})
	}

	// Copy albums list into output
	c.AlbumList2 = &AlbumList2Container{Albums: outAlbums}

	// Write response
	r.XML(200, c)
}

// GetAlbum is used in Subsonic to return a single album
func GetAlbum(req *http.Request, res http.ResponseWriter, r render.Render) {
	// Fetch ID parameter
	pID := req.URL.Query().Get("id")
	if pID == "" {
		r.XML(200, ErrMissingParameter)
		return
	}

	// Parse ID
	id, err := strconv.Atoi(pID)
	if err != nil {
		log.Println(err)
		r.XML(200, ErrGeneric)
		return
	}

	// Load album by ID
	album := &data.Album{ID: id}
	if err := album.Load(); err != nil {
		log.Println(err)
		r.XML(200, ErrGeneric)
		return
	}

	// Load songs for album
	songs, err := data.DB.SongsForAlbum(album.ID)
	if err != nil {
		log.Println(err)
		r.XML(200, ErrGeneric)
		return
	}

	// Get cover art, duration, and creation time from songs
	coverArt := 0
	duration := 0
	created := int64(0)

	outSongs := make([]Song, 0)
	for i, s := range songs {
		// Sum up duration
		duration += s.Length

		// Set cover art and created time from first song
		if i == 0 {
			coverArt = s.ArtID
			created = s.LastModified
		}

		// Build a Subsonic song
		outSongs = append(outSongs, Song{
			ID:          s.ID,
			Parent:      0,
			Title:       s.Title,
			Album:       s.Album,
			Artist:      s.Artist,
			IsDir:       false,
			CoverArt:    strconv.Itoa(coverArt),
			Created:     time.Unix(s.LastModified, 0).Format("2006-01-02T15:04:05"),
			Duration:    s.Length,
			BitRate:     s.Bitrate,
			Track:       s.Track,
			DiscNumber:  1,
			Year:        s.Year,
			Genre:       s.Genre,
			Size:        s.FileSize,
			Suffix:      "mp3",
			ContentType: "audio/mpeg",
			IsVideo:     false,
			Path:        s.FileName,
			AlbumID:     s.AlbumID,
			ArtistID:    s.ArtistID,
			Type:        "music",
		})
	}

	// Build output album
	outAlbum := &Album{
		ID:        album.ID,
		Name:      album.Title,
		Artist:    album.Artist,
		ArtistID:  album.ArtistID,
		CoverArt:  strconv.Itoa(coverArt),
		SongCount: len(songs),
		Duration:  duration,
		Created:   time.Unix(created, 0).Format("2006-01-02T15:04:05"),
	}

	// Create a new response container
	c := newContainer()

	// Copy album container into output
	outAlbum.Songs = outSongs
	c.Album = []Album{*outAlbum}

	// Write response
	r.XML(200, c)
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

// GetStream is used to return the media stream for a single file
func GetStream(req *http.Request, res http.ResponseWriter, r render.Render) {
	// Fetch ID parameter
	pID := req.URL.Query().Get("id")
	if pID == "" {
		r.XML(200, ErrMissingParameter)
		return
	}

	// Parse ID
	id, err := strconv.Atoi(pID)
	if err != nil {
		log.Println(err)
		r.XML(200, ErrGeneric)
		return
	}

	// Load song by ID
	song := &data.Song{ID: id}
	if err := song.Load(); err != nil {
		log.Println(err)
		r.XML(200, ErrGeneric)
		return
	}

	// Open file stream
	stream, err := song.Stream()
	if err != nil {
		log.Println(err)
		r.XML(200, ErrGeneric)
		return
	}

	// Generate a string used for logging this operation
	opStr := fmt.Sprintf("[#%05d] %s - %s [%s %dkbps]", song.ID, song.Artist, song.Title,
		data.CodecMap[song.FileTypeID], song.Bitrate)

	// Attempt to send file stream over HTTP
	log.Println("stream: starting:", opStr)

	// Pass stream using song's file size, auto-detect MIME type
	if err := api.HTTPStream(song, "", song.FileSize, stream, req, res); err != nil {
		// Check for client reset
		if strings.Contains(err.Error(), "connection reset by peer") || strings.Contains(err.Error(), "broken pipe") {
			return
		}

		log.Println("stream: error:", err)
		return
	}

	log.Println("stream: completed:", opStr)
	return
}
