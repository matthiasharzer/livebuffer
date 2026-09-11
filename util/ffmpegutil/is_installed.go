package ffmpegutil

import "os/exec"

func IsInstalled() bool {
	command := exec.Command("ffmpeg", "-h")
	err := command.Run()
	if err != nil {
		return false
	}
	return true
}
