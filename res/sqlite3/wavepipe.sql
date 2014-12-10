/* wavepipe sqlite3 schema */
BEGIN TRANSACTION;
/* albums */
CREATE TABLE "albums" (
	"id"          INTEGER PRIMARY KEY AUTOINCREMENT
	, "artist_id" INTEGER NOT NULL
	, "title"        TEXT NOT NULL
	, "year"      INTEGER NOT NULL
);
CREATE UNIQUE INDEX "albums_unique_artist_id_title" ON "albums" ("artist_id", "title");
/* art */
CREATE TABLE "art" (
	"id"              INTEGER PRIMARY KEY AUTOINCREMENT
	, "file_name"        TEXT NOT NULL
	, "file_size"     INTEGER NOT NULL
	, "last_modified" INTEGER NOT NULL
);
CREATE UNIQUE INDEX "art_unique_file_name" ON "art" ("file_name");
/* artists */
CREATE TABLE "artists" (
	"id"      INTEGER PRIMARY KEY AUTOINCREMENT
	, "title"    TEXT NOT NULL
);
CREATE UNIQUE INDEX "artists_unique_title" ON "artists" ("title");
/* folders */
CREATE TABLE "folders" (
	"id"          INTEGER PRIMARY KEY AUTOINCREMENT
	, "parent_id" INTEGER NOT NULL
	, "title"        TEXT NOT NULL
	, "path"         TEXT NOT NULL
);
CREATE UNIQUE INDEX "folders_unique_path" ON "folders" ("path");
/* sessions */
CREATE TABLE "sessions" (
	"id"        INTEGER PRIMARY KEY AUTOINCREMENT
	, "user_id" INTEGER NOT NULL
	, "key"        TEXT NOT NULL
	, "expire"  INTEGER NOT NULL
	, "client"     TEXT NOT NULL

	, FOREIGN KEY(user_id) REFERENCES users(id)
);
CREATE UNIQUE INDEX "sessions_unique_key" ON "sessions" ("key");
/* songs */
CREATE TABLE "songs" (
	"id"            INTEGER PRIMARY KEY AUTOINCREMENT
	, "album_id"      INTEGER NOT NULL
	, "art_id"        INTEGER NOT NULL
	, "artist_id"     INTEGER NOT NULL
	, "bitrate"       INTEGER NOT NULL
	, "channels"      INTEGER NOT NULL
	, "comment"          TEXT NOT NULL
	, "duration"      INTEGER NOT NULL
	, "file_name"        TEXT NOT NULL
	, "file_size"     INTEGER NOT NULL
	, "file_type_id"  INTEGER NOT NULL
	, "folder_id"     INTEGER NOT NULL
	, "genre"            TEXT NOT NULL
	, "last_modified" INTEGER NOT NULL
	, "sample_rate"   INTEGER NOT NULL
	, "title"            TEXT NOT NULL
	, "track"         INTEGER NOT NULL
	, "year"          INTEGER NOT NULL

	, FOREIGN KEY(album_id) REFERENCES albums(id)
	, FOREIGN KEY(art_id) REFERENCES art(id)
	, FOREIGN KEY(artist_id) REFERENCES artists(id)
	, FOREIGN KEY(folder_id) REFERENCES folders(id)
);
CREATE UNIQUE INDEX "songs_unique_file_name" ON "songs" ("file_name");
/* users */
CREATE TABLE "users" (
	"id"             INTEGER PRIMARY KEY AUTOINCREMENT
	, "username"        TEXT NOT NULL
	, "password"        TEXT NOT NULL
	, "role_id"      INTEGER NOT NULL
	, "lastfm_token"    TEXT NOT NULL
);
CREATE UNIQUE INDEX "users_unique_username" ON "users" ("username");
CREATE UNIQUE INDEX "users_unique_password" ON "users" ("password");
COMMIT;
