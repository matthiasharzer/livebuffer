package run

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/matthiasharzer/livebuffer/buffer"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "run",
	Short: "Run the livebuffer server for twitch",
	RunE: func(cmd *cobra.Command, args []string) error {

		bufferDirectory := ".buffer"
		err := os.MkdirAll(bufferDirectory, 0777)
		if err != nil {
			return err
		}

		director, err := buffer.NewDirector(2, bufferDirectory, "lars_tm")
		if err != nil {
			return err
		}

		director.WentLive()
		defer director.Close()

		fmt.Printf("Waiting for streams to be available...\n")

		time.Sleep(10 * time.Second)

		streams, err := director.GetStreams()
		if err != nil {
			return err
		}
		fmt.Printf("Streams: %v\n", streams)

		if len(streams) == 0 {
			return fmt.Errorf("no streams found")
		}

		fmt.Printf("Downloading stream: %s\n", streams[0])

		reader, err := director.GetStream(streams[0])
		if err != nil {
			return err
		}
		defer reader.Close()

		fmt.Printf("Writing stream to output.ts\n")

		fi, err := os.Create("output.ts")
		if err != nil {
			return err
		}
		defer fi.Close()

		fmt.Printf("Copying stream to output.ts\n")

		_, err = io.Copy(fi, reader)
		if err != nil {
			return err
		}

		fmt.Printf("Finished writing stream to output.ts\n")

		return nil
	},
}
