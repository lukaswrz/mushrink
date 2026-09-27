package transcoding

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func Transcode(from, to string) (err error) {
	err = os.MkdirAll(filepath.Dir(to), 0777)
	if err != nil {
		return
	}

	cmd := exec.Command(
		"ffmpeg",
		"-protocol_whitelist", // Only access the local file system.
		"file",
		"-n", // Don't overwrite output files.
		"-i",
		from,
		"-c:a",
		"libmp3lame",
		"-q:a",
		"2",
		"--",
		to,
	)

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("ffmpeg failed: %w", err)
		return
	}

	return
}
