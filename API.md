API
===

wavepipe features a simple API which is used to retrieve metadata from media files, as well as endpoints
to retrieve a file stream from the server.

An information endpoint can be found at the root of the API, `/api`.  This endpoint contains API metadata
such as the current API version, a link to this documentation, and a list of all currently available API endpoints.

At this time, the current API version is **v0**.  This API is **unstable**, and is subject to change.

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

**Return JSON:**

| Name | Type | Description |
| :--: | :--: | :---------: |
| error | [Error](http://godoc.org/github.com/mdlayher/wavepipe/api#Error)/null | Information about any errors that occurred.  Value is null if no error occurred. |
| session | string | Session ID for use with the API. On login failure, it is an empty string: "" |

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
