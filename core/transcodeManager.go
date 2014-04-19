package core

import (
	"io/ioutil"
	"log"
	"os/exec"
	"strings"

	"github.com/mdlayher/wavepipe/transcode"
)

// transcodeManager manages active file transcodes, their caching, etc, and communicates back
// and forth with the manager goroutine
func transcodeManager(transcodeKillChan chan struct{}) {
	log.Println("transcode: starting...")

	// Disable transcoding until ffmpeg is available
	transcode.Enabled = false

	// Verify that ffmpeg is available for transcoding
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		log.Println("transcode: cannot find ffmpeg, transcoding will be disabled")
		close(transcodeKillChan)
		return
	}

	// Set ffmpeg location, enable transcoding
	log.Println("transcode: found ffmpeg:", path)
	transcode.Enabled = true
	transcode.FFmpegPath = path

	// Check for codecs which wavepipe uses that ffmpeg is able to use
	ffmpeg := exec.Command(path, "-loglevel", "quiet", "-codecs")
	stdout, err := ffmpeg.StdoutPipe()

	// Start ffmpeg to retrieve its codecs, wait for it to finish
	err2 := ffmpeg.Start()
	codecs, err3 := ioutil.ReadAll(stdout)
	defer stdout.Close()
	err4 := ffmpeg.Wait()

	// Check errors
	if err != nil || err2 != nil || err3 != nil || err4 != nil {
		log.Println("transcode: could not detect ffmpeg codecs, transcoding will be disabled")
		close(transcodeKillChan)
		return
	}

	// Check for libmp3lame, for MP3 transcoding
	codecStr := string(codecs)
	if strings.Contains(codecStr, "libmp3lame MP3 (MPEG audio layer 3)") {
		log.Println("transcode: found libmp3lame, enabling MP3 transcoding")
		transcode.MP3Enabled = true
	}

	// Check for libvorbis, for OGG transcoding
	if strings.Contains(codecStr, "libvorbis Vorbis") {
		log.Println("transcode: found libvorbis, enabling OGG transcoding")
		transcode.OGGEnabled = true
	}

	transcode.MP3Enabled = false

	// Trigger events via channel
	for {
		select {
		// Stop transcode manager
		case <-transcodeKillChan:
			// Inform manager that shutdown is complete
			log.Println("transcode: stopped!")
			transcodeKillChan <- struct{}{}
			return
		}
	}
}
