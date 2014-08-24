package common

import (
	"strings"
	"time"
)

// ExpandHomeDir replaces input tilde characters with the absolute path to the current
// user's home directory.
func ExpandHomeDir(path string) string {
	return strings.Replace(path, "~", System.User.HomeDir, -1)
}

// UNIXtoRFC1123 transforms an input UNIX timestamp into the form specified by RFC1123,
// using the GMT time zone. This function is used to output Last-Modified headers via HTTP.
func UNIXtoRFC1123(unix int64) string {
	return strings.Replace(time.Unix(unix, 0).UTC().Format(time.RFC1123), "UTC", "GMT", 1)
}
