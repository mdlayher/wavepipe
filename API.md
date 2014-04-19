API
===

wavepipe features a simple API which is used to retrieve metadata from media files, as well as endpoints
to retrieve a file stream from the server.

An information endpoint can be found at the root of the API, `/api`.  This endpoint contains API metadata
such as the current API version, a link to this documentation, and a list of all currently available API endpoints.

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
		"publicKey": "abcdef0123456789abcdef0123456789",
		"secretKey": "0123456789abcdef0123456789abcdef"
	}
}
```

Upon successful login, a public key and secret key are generated, which are used to perform HMAC-SHA1
authentication with all other API endpoints.  An API signature must be calculated and included with
each request, as well as the user's public key.  The pseudocode structure of this signature is as follows:

```
signature = hmac_sha1("publicKey-nonce-method-resource", secretKey);
```

Each parameter is described as follows:
  - signature: the final resulting signature of the HMAC-SHA1 algorithm.
  - publicKey: the public key retrieved from the Login API request.
  - nonce: a randomly generated value, used only once. Repeated requests will fail.
  - method: the HTTP method used to access the resource. Typically `GET`.
  - resource: the HTTP resource being accessed, such as `/api/v0/albums`.
  - secretKey: the secret key retrieved from the Login API request. Used to validate the signature server-side.

Once the signature has been generated, it may be sent to the API using either HTTP Basic or via query string.
Both methods can be demonstrated with `curl` as follows:

```
$ curl -u publicKey:nonce:signature http://localhost:8080/api/v0/albums
$ curl http://localhost:8080/api/v0/albums?s=publicKey:nonce:signature
```

**Table of Contents:**

| Name | Versions | Description |
| :--: | :------: | :---------: |
| [Albums](#albums) | v0 | Used to retrieve information about albums from wavepipe. |
| [Artists](#artists) | v0 | Used to retrieve information about artists from wavepipe. |
| [Folders](#folders) | v0 | Used to retrieve information about folders from wavepipe. |
| [Login](#login) | v0 | Used to generate a new API session on wavepipe. |
| [Logout](#logout) | v0 | Used to destroy the current API session from wavepipe. |
| [Songs](#songs) | v0 | Used to retrieve information about songs from wavepipe. |
| [Stream](#stream) | v0 | Used to retrieve a raw, non-transcoded, binary data stream of a media file from wavepipe. |
| [Transcode](#transcode) | v0 | Used to retrieve transcoded binary data stream of a media file from wavepipe. |

## Albums
Used to retrieve information about albums from wavepipe.  If an ID is specified, information will be
retrieved about a single album.

**Versions:** `v0`

**URL:** `/api/v0/albums/:id`

**Examples:** `http://localhost:8080/api/v0/albums/`, `http://localhost:8080/api/v0/albums/1`

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
| 400 | invalid integer album ID | A valid integer could not be parsed from the ID. |
| 404 | album ID not found | An album with the specified ID does not exist. |
| 500 | server error | An internal error occurred. wavepipe will log these errors to its console log. |

## Artists
Used to retrieve information about artists from wavepipe.  If an ID is specified, information will be
retrieved about a single artist.

**Versions:** `v0`

**URL:** `/api/v0/artists/:id`

**Examples:** `http://localhost:8080/api/v0/artists/`, `http://localhost:8080/api/v0/artists/1`, `http://localhost:8080/api/v0/artists/1?songs=true`

**Parameters:**

| Name | Versions | Type | Required | Description |
| :--: | :------: | :--: | :------: | :---------: |
| songs | v0 | bool | | If true, returns all songs attached to this artist. |

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
| 400 | invalid integer artist ID | A valid integer could not be parsed from the ID. |
| 404 | artist ID not found | An artist with the specified ID does not exist. |
| 500 | server error | An internal error occurred. wavepipe will log these errors to its console log. |

## Folders
Used to retrieve information about folders from wavepipe.  If an ID is specified, information will be
retrieved about a single folder.

**Versions:** `v0`

**URL:** `/api/v0/folders/:id`

**Examples:** `http://localhost:8080/api/v0/folders/`, `http://localhost:8080/api/v0/folders/1`

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
| 400 | invalid integer folder ID | A valid integer could not be parsed from the ID. |
| 404 | folder ID not found | An folder with the specified ID does not exist. |
| 500 | server error | An internal error occurred. wavepipe will log these errors to its console log. |

## Login
Used to generate a new API session on wavepipe.  Credentials may be provided either via query string,
or using a HTTP Basic username and password combination.

**Versions:** `v0`

**URL:** `/api/v0/login`

**Examples:** `http://localhost:8080/api/v0/login`

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

**Examples:** `http://localhost:8080/api/v0/logout`

**Return JSON:**

| Name | Type | Description |
| :--: | :--: | :---------: |
| error | [Error](http://godoc.org/github.com/mdlayher/wavepipe/api#Error)/null | Information about any errors that occurred.  Value is null if no error occurred. |

**Possible errors:**

| Code | Message | Description |
| :--: | :-----: | :---------: |
| 400 | unsupported API version: vX | Attempted access to an invalid version of this API, or to a version before this API existed. |
| 500 | server error | An internal error occurred. wavepipe will log these errors to its console log. |

## Songs
Used to retrieve information about songs from wavepipe.  If an ID is specified, information will be
retrieved about a single song.

**Versions:** `v0`

**URL:** `/api/v0/songs/:id`

**Examples:** `http://localhost:8080/api/v0/songs/`, `http://localhost:8080/api/v0/songs/1`

**Return JSON:**

| Name | Type | Description |
| :--: | :--: | :---------: |
| error | [Error](http://godoc.org/github.com/mdlayher/wavepipe/api#Error)/null | Information about any errors that occurred.  Value is null if no error occurred. |
| songs | \[\][Song](http://godoc.org/github.com/mdlayher/wavepipe/data#Song) | Array of Song objects returned by the API. |

**Possible errors:**

| Code | Message | Description |
| :--: | :-----: | :---------: |
| 400 | unsupported API version: vX | Attempted access to an invalid version of this API, or to a version before this API existed. |
| 400 | invalid integer song ID | A valid integer could not be parsed from the ID. |
| 404 | song ID not found | A song with the specified ID does not exist. |
| 500 | server error | An internal error occurred. wavepipe will log these errors to its console log. |

## Stream
Used to retrieve a raw, non-transcoded, binary data stream of a media file from wavepipe.  An ID **must** be specified to access a file stream.  Successful calls with return a binary stream, and unsuccessful ones will return a JSON error.

**Versions:** `v0`

**URL:** `/api/v0/stream/:id`

**Examples:** `http://localhost:8080/api/v0/stream/1`

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

**Examples:** `http://localhost:8080/api/v0/transcode/1`, `http://localhost:8080/api/v0/transcode/1?codec=MP3&quality=320`

**Parameters:**

| Name | Versions | Type | Required | Description |
| :--: | :------: | :--: | :------: | :---------: |
| codec | v0 | string | | The codec selected for use by the transcoder.  If not specified, defaults to **MP3**.  Options are: **MP3**, OGG. |
| quality | v0 | string/int | | The quality selected for use by the transcoder.  String options specify VBR encodings, while integer options specify CBR encodings.  If not specified, defaults to **192**. |

**Available Codecs:**

| Codec | Versions | Type | Options | Description |
| :---: | :------: | :--: | :-----: | :---------: |
| MP3 | v0 | CBR | 128, **192** (default), 256, 320 | Generates a constant bitrate encode using LAME. |
| MP3 | v0 | VBR | V0 (~245kbps), V2 (~190kbps), V4 (~165kbps) | Generates a variable bitrate encode using a specific LAME quality level. |
| OGG | v0 | CBR | 128, **192** (default), 256, 320, 500 | Generates a constant bitrate encode using Ogg Vorbis. |
| OGG | v0 | VBR | Q10 (~500kbps), Q8 (~256kbps), Q6 (~192kbps) | Generates a variable bitrate encode using a specific Ogg Vorbis quality level. |

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
