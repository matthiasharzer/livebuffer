package main

import (
	"fmt"
	"os"

	"github.com/matthiasharzer/livebuffer/cmd/twitch"
	"github.com/matthiasharzer/livebuffer/cmd/version"

	"github.com/spf13/cobra"
)

var rootCommand = &cobra.Command{
	Use:   "livebuffer",
	Short: "livebuffer is a tool to buffer live streams from Twitch and provide a public URL to access the buffered content",
	Long:  "livebuffer is a tool to buffer live streams from Twitch and provide a public URL to access the buffered content. It allows you to specify a Twitch username, and it will automatically record the live stream when the user goes live. The recorded segments are stored in a specified directory, and you can access the buffered content through a public URL",
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCommand.AddCommand(version.Command)
	rootCommand.AddCommand(twitch.Command)
}

func main() {
	err := rootCommand.Execute()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
