API
===

wavepipe features a simple API which is used to retrieve metadata from media files, as well as endpoints
to retrieve a file stream from the server.

At this time, the API is **unstable**, and is subject to change.

## Albums
Used to retrieve information about albums from wavepipe.  If an ID is specified, information will be
retrieved about a single album.

**URL:** `/api/albums/:id`

**Examples:** `http://localhost:8080/api/albums/`, `http://localhost:8080/api/albums/1`

**Return JSON:**

| Name | Type | Description |
| :--: | :--: | :---------: |
| error | Error/null | Information about any errors that occurred.  Value is null if no error occurred. |
| albums | []Album | Array of Album objects returned by the API. |
| songs | []Song/null | If ID is specified, array of Song objects attached to this album.  Value is null if no ID specified. |
