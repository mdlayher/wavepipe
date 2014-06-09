API
===

wavepipe features a simple API which is used to retrieve metadata from media files, as well as endpoints
to retrieve a file stream from the server.

An information endpoint can be found at the root of the API, `/api`.  This endpoint contains API metadata
such as the current API version, supported API versions, and a link to this documentation.

At this time, the current API version is **v0**.  This API is **unstable**, and is subject to change.

Unless otherwise noted, API endpoints should be accessed using the HTTP `GET` method, due to the vast majority
of API endpoints being read-only.

**Authentication:**

In order to use the wavepipe API, all requests must be authenticated.  The first step is to generate a new
session via the [Login](#login) API.  Login username and password can be passed using either HTTP Basic or
via query string.  In addition, an optional client parameter may be passed, which will identify this session
with the given name. Both methods can be demonstrated with `curl` as follows:

```
$ curl -u test:test http://localhost:8080/api/v0/login?c=testclient
$ curl http://localhost:8080/api/v0/login?u=test&p=test&c=testclient
```

Example [Session](http://godoc.org/github.com/mdlayher/wavepipe/data#Session) output is as follows:

```json
{
	"error": null,
	"session": {
		"id": 1,
		"userId": 1,
		"client": "testclient",
		"expire": 1397713157,
		"key": "abcdef0123456789abcdef0123456789"
	}
}
```

Upon successful login, a session key is generated, which is used to authenticate subsequent requests.
It should be noted that unless the service is secured via HTTPS, this token can be compromised by other
users on the same network.  For this reason, it is recommended to place wavepipe behind SSL.

This method can be demonstrated with `curl` as follows.

```
$ curl http://localhost:8080/api/v0/albums?s=abcdef0123456789abcdef0123456789
```

Sessions which are not used for one week will expire.  Each subsequent API request with a specified session will
update the expiration time to one week in the future.

**Table of Contents:**

| Name | Versions | Description |
| :--: | :------: | :---------: |
| [Albums](#albums) | v0 | Used to retrieve information about albums from wavepipe. |
| [Art](#art) | v0 | Used to retrieve a binary data stream of an art file from wavepipe. |
| [Artists](#artists) | v0 | Used to retrieve information about artists from wavepipe. |
| [Folders](#folders) | v0 | Used to retrieve information about folders from wavepipe. |
| [LastFM](#lastfm) | v0 | Used to scrobble songs from wavepipe to Last.fm. |
| [Login](#login) | v0 | Used to generate a new API session on wavepipe. |
| [Logout](#logout) | v0 | Used to destroy the current API session from wavepipe. |
| [Search](#search) | v0 | Used to retrieve artists, albums, songs, and folders which match a specified search query. |
| [Songs](#songs) | v0 | Used to retrieve information about songs from wavepipe. |
| [Status](#status) | v0 | Used to retrieve current server status from wavepipe. |
| [Stream](#stream) | v0 | Used to retrieve a raw, non-transcoded, binary data stream of a media file from wavepipe. |
| [Transcode](#transcode) | v0 | Used to retrieve transcoded binary data stream of a media file from wavepipe. |

## Albums
Used to retrieve information about albums from wavepipe.  If an ID is specified, information will be
retrieved about a single album.

**Versions:** `v0`

**URL:** `/api/v0/albums/:id`

**Examples:**
  - `http://localhost:8080/api/v0/albums/`
  - `http://localhost:8080/api/v0/albums/1`
  - `http://localhost:8080/api/v0/albums?limit=0,100`

**Parameters:**

| Name | Versions | Type | Required | Description |
| :--: | :------: | :--: | :------: | :---------: |
| limit | v0 | integer,integer | | Comma-separated integer pair which limits the number of returned results.  First integer is the offset, second integer is the item count. |

**Return JSON:**

| Name | Type | Description |
| :--: | :--: | :---------: |
| error | [Error](http://godoc.org/github.com/mdlayher/wavepipe/api#Error)/null | Information about any errors that occurred.  Value is null if no error occurred. |
| albums | \[\][Album](http://godoc.org/github.com/mdlayher/wavepipe/data#Album) | Array of Album objects returned by the API. |
| songs | \[\][Song](http://godoc.org/github.com/mdlayher/wavepipe/data#Song)/null | If ID is specified, array of Song objects attached to this album.  Value is null if no ID specified. |

**Possible errors:**

| Code | Message | Description |
| :--: | :-----: | :---------: |
| 400 | unsupported API version: vX | Attempted access to an invalid version of this API, or to a version before this API existed. |
| 400 | invalid comma-separated integer pair for limit | A valid integer pair could not be parsed from the limit parameter. Input must be in the form "x,y". |
| 400 | invalid integer album ID | A valid integer could not be parsed from the ID. |
| 404 | album ID not found | An album with the specified ID does not exist. |
| 500 | server error | An internal error occurred. wavepipe will log these errors to its console log. |

## Art
Used to retrieve a binary data stream of an art file from wavepipe.  An ID **must** be specified to access an art stream.
Successful calls with return a binary stream, and unsuccessful ones will return a JSON error.

**Versions:** `v0`

**URL:** `/api/v0/art/:id`

**Examples:**
  - `http://localhost:8080/api/v0/art/1`
  - `http://localhost:8080/api/v0/art/1?size=500`

**Parameters:**

| Name | Versions | Type | Required | Description |
| :--: | :------: | :--: | :------: | :---------: |
| size | v0 | integer | | Scale the art to the specified size in pixels. The art's original aspect ratio will be preserved. |

**Return Binary:** Binary data stream containing the art file stream.

**Return JSON:**

| Name | Type | Description |
| :--: | :--: | :---------: |
| error | [Error](http://godoc.org/github.com/mdlayher/wavepipe/api#Error) | Information about any errors that occurred. |

**Possible errors:**

| Code | Message | Description |
| :--: | :-----: | :---------: |
| 400 | unsupported API version: vX | Attempted access to an invalid version of this API, or to a version before this API existed. |
| 400 | no integer art ID provided | No integer ID was sent in request. |
| 400 | invalid art stream ID | A valid integer could not be parsed from the ID. |
| 400 | negative integer size | A negative integer was passed to the size parameter. Size **must** be a positive integer. |
| 404 | art ID not found | An art file with the specified ID does not exist. |
| 500 | server error | An internal error occurred. wavepipe will log these errors to its console log. |

## Artists
Used to retrieve information about artists from wavepipe.  If an ID is specified, information will be
retrieved about a single artist.

**Versions:** `v0`

**URL:** `/api/v0/artists/:id`

**Examples:**
  - `http://localhost:8080/api/v0/artists/`
  - `http://localhost:8080/api/v0/artists/1`
  - `http://localhost:8080/api/v0/artists?limit=0,100`
  - `http://localhost:8080/api/v0/artists/1?songs=true`

**Parameters:**

| Name | Versions | Type | Required | Description |
| :--: | :------: | :--: | :------: | :---------: |
| limit | v0 | string "integer,integer" | | Comma-separated integer pair which limits the number of returned results.  First integer is the offset, second integer is the item count. |
| songs | v0 | boolean | | If true, returns all songs attached to this artist. |

**Return JSON:**

| Name | Type | Description |
| :--: | :--: | :---------: |
| error | [Error](http://godoc.org/github.com/mdlayher/wavepipe/api#Error)/null | Information about any errors that occurred.  Value is null if no error occurred. |
| artists | \[\][Artist](http://godoc.org/github.com/mdlayher/wavepipe/data#Artist) | Array of Artist objects returned by the API. |
| albums | \[\][Album](http://godoc.org/github.com/mdlayher/wavepipe/data#Album)/null | If ID is specified, array of Album objects attached to this artist. |
| songs | \[\][Song](http://godoc.org/github.com/mdlayher/wavepipe/data#Song)/null | If parameter `songs` is true, array of Song objects attached to this artist.  Value is null if parameter `songs` is false or not specified. |

**Possible errors:**

| Code | Message | Description |
| :--: | :-----: | :---------: |
| 400 | unsupported API version: vX | Attempted access to an invalid version of this API, or to a version before this API existed. |
| 400 | invalid comma-separated integer pair for limit | A valid integer pair could not be parsed from the limit parameter. Input must be in the form "x,y". |
| 400 | invalid integer artist ID | A valid integer could not be parsed from the ID. |
| 404 | artist ID not found | An artist with the specified ID does not exist. |
| 500 | server error | An internal error occurred. wavepipe will log these errors to its console log. |

## Folders
Used to retrieve information about folders from wavepipe.  If an ID is specified, information will be
retrieved about a single folder.

**Versions:** `v0`

**URL:** `/api/v0/folders/:id`

**Examples:**
  - `http://localhost:8080/api/v0/folders/`
  - `http://localhost:8080/api/v0/folders/1`
  - `http://localhost:8080/api/v0/folders?limit=0,100`

**Parameters:**

| Name | Versions | Type | Required | Description |
| :--: | :------: | :--: | :------: | :---------: |
| limit | v0 | integer,integer | | Comma-separated integer pair which limits the number of returned results.  First integer is the offset, second integer is the item count. |

**Return JSON:**

| Name | Type | Description |
| :--: | :--: | :---------: |
| error | [Error](http://godoc.org/github.com/mdlayher/wavepipe/api#Error)/null | Information about any errors that occurred.  Value is null if no error occurred. |
| folders | \[\][Folder](http://godoc.org/github.com/mdlayher/wavepipe/data#Folder) | Array of Folder objects returned by the API. |
| subfolders | \[\][Folder](http://godoc.org/github.com/mdlayher/wavepipe/data#Folder) | If ID is specified, array of Folder objects which are children to the current folder, returned by the API. Value is null if no ID is specified. |
| songs | \[\][Song](http://godoc.org/github.com/mdlayher/wavepipe/data#Song)/null | If ID is specified, array of Song objects attached to this folder.  Value is null if no ID specified. |

**Possible errors:**

| Code | Message | Description |
| :--: | :-----: | :---------: |
| 400 | unsupported API version: vX | Attempted access to an invalid version of this API, or to a version before this API existed. |
| 400 | invalid comma-separated integer pair for limit | A valid integer pair could not be parsed from the limit parameter. Input must be in the form "x,y". |
| 400 | invalid integer folder ID | A valid integer could not be parsed from the ID. |
| 404 | folder ID not found | An folder with the specified ID does not exist. |
| 500 | server error | An internal error occurred. wavepipe will log these errors to its console log. |

## LastFM
Used to scrobble songs from wavepipe to Last.fm.  The user must first complete a `login` action with their Last.fm
credentials, and then the `nowplaying` and `scrobble` actions may be used.  After the initial `login`, wavepipe
will store an API key for the user, and use this key for future requests.

Ideally, a `nowplaying` action will be triggered by clients as soon as the track begins playing on that client.
After a fair amount of time has passed (for example, 50-75% of the song), a `scrobble` request should be triggered
to commit the play to Last.fm.

**Versions:** `v0`

**URL:** `/api/v0/lastfm/:action/:id`

**Examples:**
  - `http://localhost:8080/api/v0/lastfm/login?lfmu=test&lfmp=test`
  - `http://localhost:8080/api/v0/lastfm/nowplaying/1`
  - `http://localhost:8080/api/v0/lastfm/scrobble/1`

**Parameters:**

| Name | Versions | Type | Required | Description |
| :--: | :------: | :--: | :------: | :---------: |
| lfmu | v0 | string | | Username used to authenticate to Last.fm via wavepipe. Only used for the `login` action. |
| lfmp | v0 | string | | Associated password used to authenticate to Last.fm via wavepipe. Only used for the `login` action. |
| timestamp | v0 | integer | | Optional integer UNIX timestamp, which can be used to specify a past timestamp. The current timestamp is used if not specified. |

**Return JSON:**

| Name | Type | Description |
| :--: | :--: | :---------: |
| error | [Error](http://godoc.org/github.com/mdlayher/wavepipe/api#Error)/null | Information about any errors that occurred.  Value is null if no error occurred. |
| url | string | String containing the URL required to authorize wavepipe's Last.fm token for this user. Only returned on the `login` action, or if other actions are accessed while the token is not authorized. |

**Possible errors:**

| Code | Message | Description |
| :--: | :-----: | :---------: |
| 400 | unsupported API version: vX | Attempted access to an invalid version of this API, or to a version before this API existed. |
| 400 | no string action provided | No action was specified in the URL.  An action **must** be specified to use Last.fm functionality. |
| 400 | invalid string action provided | An unknown action was specified in the URL.  Valid actions are `login`, `nowplaying`, and `scrobble`. |
| 400 | login: no username provided | No Last.fm username was passed via `lfmu` in query string. Only returned on `login` action. |
| 400 | login: no password provided | No Last.fm password was passed via `lfmp` in query string. Only returned on `login` action. |
| 400 | no integer song ID provided | No integer ID was sent in request. Only returned on `nowplaying` and `scrobble` actions. |
| 400 | invalid integer song ID | A valid integer could not be parsed from the ID. Only returned on `nowplaying` and `scrobble` actions. |
| 401 | action: last.fm authentication failed | Could not authenticate to Last.fm. Could be due to invalid username/password, or an invalid API token. |
| 401 | action: user must authenticate to last.fm | User attempted to perform `nowplaying` or `scrobble` action, without first completing `login` action. |
| 401 | action: last.fm token not yet authorized | User must authorize wavepipe to access their Last.fm account, via the provided URL. |
| 404 | song ID not found | A song with the specified ID does not exist. Only returned on `nowplaying` and `scrobble` actions. |
| 500 | server error | An internal error occurred. wavepipe will log these errors to its console log. |

## Login
Used to generate a new API session on wavepipe.  Credentials may be provided either via query string,
or using a HTTP Basic username and password combination.

**Versions:** `v0`

**URL:** `/api/v0/login`

**Examples:**
  - `http://localhost:8080/api/v0/login`

**Parameters:**

| Name | Versions | Type | Required | Description |
| :--: | :------: | :--: | :------: | :---------: |
| u | v0 | string | X | Username used to authenticate to wavepipe. Can also be passed via HTTP Basic. |
| p | v0 | string | X | Associated password used to authenticate to wavepipe. Can also be passed via HTTP Basic. |
| c | v0 | string | | Optional client name used to identify this session. |

**Return JSON:**

| Name | Type | Description |
| :--: | :--: | :---------: |
| error | [Error](http://godoc.org/github.com/mdlayher/wavepipe/api#Error)/null | Information about any errors that occurred.  Value is null if no error occurred. |
| session | [Session](http://godoc.org/github.com/mdlayher/wavepipe/data#Session) | Session object which contains the public and secret keys used to authenticate further API calls. |

**Possible errors:**

| Code | Message | Description |
| :--: | :-----: | :---------: |
| 400 | unsupported API version: vX | Attempted access to an invalid version of this API, or to a version before this API existed. |
| 401 | authentication failed: X | API authentication failed. Could be due to malformed, missing, or bad credentials. |
| 500 | server error | An internal error occurred. wavepipe will log these errors to its console log. |

## Logout
Used to destroy the current API session from wavepipe.

**Versions:** `v0`

**URL:** `/api/v0/logout`

**Examples:**
  - `http://localhost:8080/api/v0/logout`

**Return JSON:**

| Name | Type | Description |
| :--: | :--: | :---------: |
| error | [Error](http://godoc.org/github.com/mdlayher/wavepipe/api#Error)/null | Information about any errors that occurred.  Value is null if no error occurred. |

**Possible errors:**

| Code | Message | Description |
| :--: | :-----: | :---------: |
| 400 | unsupported API version: vX | Attempted access to an invalid version of this API, or to a version before this API existed. |
| 500 | server error | An internal error occurred. wavepipe will log these errors to its console log. |

## Search
Used to retrieve artists, albums, songs, and folders which match a specified search query.  A search query **must** be
specified to retrieve results.

**Versions:** `v0`

**URL:** `/api/v0/search/:query`

**Examples:**
  - `http://localhost:8080/api/v0/search/boston`
  - `http://localhost:8080/api/v0/search/boston?type=artists,songs`

**Parameters:**

| Name | Versions | Type | Required | Description |
| :--: | :------: | :--: | :------: | :---------: |
| type | v0 | string | | Comma-separated string containing object types (`artists`, `albums`, `songs`, `folders`) to return search results. If not specified, equivalent to `artists,albums,songs,folders`. |

**Return JSON:**

| Name | Type | Description |
| :--: | :--: | :---------: |
| error | [Error](http://godoc.org/github.com/mdlayher/wavepipe/api#Error) | Information about any errors that occurred. |
| artists | \[\][Artist](http://godoc.org/github.com/mdlayher/wavepipe/data#Artist) | Array of Artist objects with titles matching the search query. |
| albums | \[\][Album](http://godoc.org/github.com/mdlayher/wavepipe/data#Album) | Array of Album objects with titles matching the search query. |
| songs | \[\][Song](http://godoc.org/github.com/mdlayher/wavepipe/data#Song) | Array of Song objects with titles matching the search query. |
| folders | \[\][Folder](http://godoc.org/github.com/mdlayher/wavepipe/data#Folder) | Array of Folder objects with titles matching the search query. |

**Possible errors:**

| Code | Message | Description |
| :--: | :-----: | :---------: |
| 400 | unsupported API version: vX | Attempted access to an invalid version of this API, or to a version before this API existed. |
| 400 | no search query specified | No search query was specified in the URL. A search query **must** be specified to retrieve results. |
| 500 | server error | An internal error occurred. wavepipe will log these errors to its console log. |

## Songs
Used to retrieve information about songs from wavepipe.  If an ID is specified, information will be
retrieved about a single song.

**Versions:** `v0`

**URL:** `/api/v0/songs/:id`

**Examples:**
  - `http://localhost:8080/api/v0/songs/`
  - `http://localhost:8080/api/v0/songs/1`
  - `http://localhost:8080/api/v0/songs?limit=0,100`
  - `http://localhost:8080/api/v0/songs?random=10`

**Parameters:**

| Name | Versions | Type | Required | Description |
| :--: | :------: | :--: | :------: | :---------: |
| limit | v0 | integer,integer | | Comma-separated integer pair which limits the number of returned results.  First integer is the offset, second integer is the item count. |
| random | v0 | integer | | If specified, wavepipe will return N random songs instead of the entire list, where N is the integer specified in this parameter. |

**Return JSON:**

| Name | Type | Description |
| :--: | :--: | :---------: |
| error | [Error](http://godoc.org/github.com/mdlayher/wavepipe/api#Error)/null | Information about any errors that occurred.  Value is null if no error occurred. |
| songs | \[\][Song](http://godoc.org/github.com/mdlayher/wavepipe/data#Song) | Array of Song objects returned by the API. |

**Possible errors:**

| Code | Message | Description |
| :--: | :-----: | :---------: |
| 400 | unsupported API version: vX | Attempted access to an invalid version of this API, or to a version before this API existed. |
| 400 | invalid comma-separated integer pair for limit | A valid integer pair could not be parsed from the limit parameter. Input must be in the form "x,y". |
| 400 | invalid integer for random | A valid integer could not be parsed from the random parameter. |
| 400 | invalid integer song ID | A valid integer could not be parsed from the ID. |
| 404 | song ID not found | A song with the specified ID does not exist. |
| 500 | server error | An internal error occurred. wavepipe will log these errors to its console log. |

## Status
Used to retrieve current server status from wavepipe.

**Versions:** `v0`

**URL:** `/api/v0/status`

**Examples:**
  - `http://localhost:8080/api/v0/status`
  - `http://localhost:8080/api/v0/status?metrics=true`

**Parameters:**

| Name | Versions | Type | Required | Description |
| :--: | :------: | :--: | :------: | :---------: |
| metrics | v0 | boolean | | If true, wavepipe will generate and return additional metrics about its database. |

**Return JSON:**

| Name | Type | Description |
| :--: | :--: | :---------: |
| error | [Error](http://godoc.org/github.com/mdlayher/wavepipe/api#Error)/null | Information about any errors that occurred.  Value is null if no error occurred. |
| status | [Status](http://godoc.org/github.com/mdlayher/wavepipe/common#Status) | Status object containing current server information, returned by the API. |

**Possible errors:**

| Code | Message | Description |
| :--: | :-----: | :---------: |
| 400 | unsupported API version: vX | Attempted access to an invalid version of this API, or to a version before this API existed. |
| 500 | server error | An internal error occurred. wavepipe will log these errors to its console log. |

## Stream
Used to retrieve a raw, non-transcoded, binary data stream of a media file from wavepipe.  An ID **must** be specified to access a file stream.  Successful calls with return a binary stream, and unsuccessful ones will return a JSON error.

**Versions:** `v0`

**URL:** `/api/v0/stream/:id`

**Examples:**
  - `http://localhost:8080/api/v0/stream/1`

**Return Binary:** Binary data stream containing the media file stream.

**Return JSON:**

| Name | Type | Description |
| :--: | :--: | :---------: |
| error | [Error](http://godoc.org/github.com/mdlayher/wavepipe/api#Error) | Information about any errors that occurred. |

**Possible errors:**

| Code | Message | Description |
| :--: | :-----: | :---------: |
| 400 | unsupported API version: vX | Attempted access to an invalid version of this API, or to a version before this API existed. |
| 400 | no integer stream ID provided | No integer ID was sent in request. |
| 400 | invalid integer stream ID | A valid integer could not be parsed from the ID. |
| 404 | song ID not found | A song with the specified ID does not exist. |
| 500 | server error | An internal error occurred. wavepipe will log these errors to its console log. |

## Transcode
Used to retrieve a transcoded binary data stream of a media file from wavepipe.  An ID **must** be specified to access a file stream.  Successful calls with return a binary stream, and unsuccessful ones will return a JSON error.

**Versions:** `v0`

**URL:** `/api/v0/transcode/:id`

**Examples:**
  - `http://localhost:8080/api/v0/transcode/1`
  - `http://localhost:8080/api/v0/transcode/1?codec=MP3&quality=320`

**Parameters:**

| Name | Versions | Type | Required | Description |
| :--: | :------: | :--: | :------: | :---------: |
| codec | v0 | string | | The codec selected for use by the transcoder.  If not specified, defaults to **MP3**.  Options are: **MP3**, OGG, OPUS (lowercase variants will be automatically capitalized). |
| quality | v0 | string/integer | | The quality selected for use by the transcoder.  String options specify VBR encodings, while integer options specify CBR encodings.  If not specified, defaults to **192**. |

**Available Codecs:**

| Codec | Versions | Type | Options | Description |
| :---: | :------: | :--: | :-----: | :---------: |
| MP3 | v0 | CBR | 128, **192** (default), 256, 320 | Generates a constant bitrate encode using LAME. |
| MP3 | v0 | VBR | V0 (~245kbps), V2 (~190kbps), V4 (~165kbps) | Generates a variable bitrate encode using a specific LAME quality level. |
| OGG | v0 | CBR | 128, **192** (default), 256, 320, 500 | Generates a constant bitrate encode using Ogg Vorbis. |
| OGG | v0 | VBR | Q10 (~500kbps), Q8 (~256kbps), Q6 (~192kbps) | Generates a variable bitrate encode using a specific Ogg Vorbis quality level. |
| OPUS | v0 | CBR | 128, **192** (default), 256, 320, 500 | Generates a constant bitrate encode using Ogg Opus. |
| OPUS | v0 | VBR | Q10 (~500kbps), Q8 (~256kbps), Q6 (~192kbps) | Generates a variable bitrate encode using a specific Ogg Opus quality level. |

**Return Binary:** Binary data stream containing the transcoded media file stream.

**Return JSON:**

| Name | Type | Description |
| :--: | :--: | :---------: |
| error | [Error](http://godoc.org/github.com/mdlayher/wavepipe/api#Error) | Information about any errors that occurred. |

**Possible errors:**

| Code | Message | Description |
| :--: | :-----: | :---------: |
| 400 | unsupported API version: vX | Attempted access to an invalid version of this API, or to a version before this API existed. |
| 400 | no integer transcode ID provided | No integer ID was sent in request. |
| 400 | invalid integer transcode ID | A valid integer could not be parsed from the ID. |
| 400 | invalid transcoder codec: X | A non-existant transcoder codec was passed via the codec parameter. |
| 400 | invalid quality for codec X: X | A non-existant quality setting for the specified codec was passed via the quality parameter. |
| 404 | song ID not found | A song with the specified ID does not exist. |
| 500 | server error | An internal error occurred. wavepipe will log these errors to its console log. |
| 503 | ffmpeg not found, transcoding disabled | ffmpeg binary could not be detected in system PATH, so the transcoding subsystem is disabled. |
| 503 | ffmpeg codec libmp3lame not found, MP3 transcoding disabled | ffmpeg was not compiled with libmp3lame codec, so MP3 transcoding is disabled. |
| 503 | ffmpeg codec libvorbis not found, OGG transcoding disabled | ffmpeg was not compiled with libvorbis codec, so Ogg Vorbis transcoding is disabled. |
| 503 | ffmpeg codec libopus not found, OPUS transcoding disabled | ffmpeg was not compiled with libopus codec, so Ogg Opus transcoding is disabled. |
