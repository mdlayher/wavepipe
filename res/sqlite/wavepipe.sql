/* wavepipe sqlite schema */
PRAGMA foreign_keys = OFF;
BEGIN TRANSACTION;
/* albums */
CREATE TABLE "albums" (
	"id"        INTEGER PRIMARY KEY AUTOINCREMENT,
	"artist_id" INTEGER NOT NULL,
	"title"     TEXT,
	"year"      INTEGER
);
CREATE UNIQUE INDEX "albums_unique_artistId_title" ON "albums" ("artist_id", "title");
/* artists */
CREATE TABLE "artists" (
	"id"    INTEGER PRIMARY KEY AUTOINCREMENT,
	"title" TEXT
);
CREATE UNIQUE INDEX "artists_unique_title" ON "artists" ("title");
/* songs */
CREATE TABLE "songs" (
	"id"            INTEGER PRIMARY KEY AUTOINCREMENT,
	"album_id"      INTEGER NOT NULL,
	"artist_id"     INTEGER NOT NULL,
	"bitrate"       INTEGER NOT NULL,
	"channels"      INTEGER NOT NULL,
	"comment"       TEXT,
	"file_name"     TEXT,
	"file_size"     INTEGER NOT NULL,
	"file_type"     TEXT,
	"genre"         TEXT,
	"last_modified" INTEGER NOT NULL,
	"length"        INTEGER NOT NULL,
	"sample_rate"   INTEGER NOT NULL,
	"title"         TEXT,
	"track"         INTEGER,
	"year"          INTEGER
);
CREATE UNIQUE INDEX "songs_unique_fileName_title" ON "songs" ("file_name", "title");
COMMIT;
