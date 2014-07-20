package common

import (
	"strings"
	"time"
)

// UNIXtoRFC1123 transforms an input UNIX timestamp into the form specified by RFC1123,
// using the GMT time zone. This function is used to output Last-Modified headers via HTTP.
func UNIXtoRFC1123(unix int64) string {
	return strings.Replace(time.Unix(unix, 0).UTC().Format(time.RFC1123), "UTC", "GMT", 1)
}
