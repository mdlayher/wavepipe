/* wavepipe sqlite schema */
PRAGMA foreign_keys = OFF;
PRAGMA synchronous = OFF;
BEGIN TRANSACTION;
/* albums */
CREATE TABLE "albums" (
	"artist_id" INTEGER NOT NULL,
	"title"     TEXT,
	"year"      INTEGER
);
/* artists */
CREATE TABLE "artists" (
	"title" TEXT
);
/* songs */
CREATE TABLE "songs" (
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
COMMIT;
