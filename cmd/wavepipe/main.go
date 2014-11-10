// Command wavepipe provides a cross-platform media server, written in Go.
// For usage documentation, please see the project on GitHub: https://github.com/mdlayher/wavepipe
package main

import (
	"flag"
	"log"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	flag.Parse()

	log.Println("wavepipe media server")
}
