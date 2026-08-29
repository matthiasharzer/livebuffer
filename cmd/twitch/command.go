package twitch

import (
	"github.com/matthiasharzer/livebuffer/cmd/twitch/run"
	"github.com/spf13/cobra"
)

func init() {
	Command.AddCommand(run.Command)
}

var Command = &cobra.Command{
	Use: "twitch",
}
