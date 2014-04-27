package data

// DB is the current database backend
var DB dbBackend

// dbBackend represents the database backend that the program will connect to
type dbBackend interface {
	Open() error
	Close() error
	Setup() error
	DSN(string)

	ArtInPath(string) ([]Art, error)
	ArtNotInPath(string) ([]Art, error)
	DeleteArt(*Art) error
	LoadArt(*Art) error
	SaveArt(*Art) error

	AllArtists() ([]Artist, error)
	SearchArtists(string) ([]Artist, error)
	PurgeOrphanArtists() (int, error)
	DeleteArtist(*Artist) error
	LoadArtist(*Artist) error
	SaveArtist(*Artist) error

	AllAlbums() ([]Album, error)
	AlbumsForArtist(int) ([]Album, error)
	SearchAlbums(string) ([]Album, error)
	PurgeOrphanAlbums() (int, error)
	DeleteAlbum(*Album) error
	LoadAlbum(*Album) error
	SaveAlbum(*Album) error

	AllFolders() ([]Folder, error)
	Subfolders(int) ([]Folder, error)
	FoldersInPath(string) ([]Folder, error)
	FoldersNotInPath(string) ([]Folder, error)
	SearchFolders(string) ([]Folder, error)
	DeleteFolder(*Folder) error
	LoadFolder(*Folder) error
	SaveFolder(*Folder) error

	AllSongs() ([]Song, error)
	SearchSongs(string) ([]Song, error)
	SongsForAlbum(int) ([]Song, error)
	SongsForArtist(int) ([]Song, error)
	SongsForFolder(int) ([]Song, error)
	SongsInPath(string) ([]Song, error)
	SongsNotInPath(string) ([]Song, error)
	DeleteSong(*Song) error
	LoadSong(*Song) error
	SaveSong(*Song) error
	UpdateSong(*Song) error

	DeleteUser(*User) error
	LoadUser(*User) error
	SaveUser(*User) error

	DeleteSession(*Session) error
	LoadSession(*Session) error
	SaveSession(*Session) error
	UpdateSession(*Session) error
}
