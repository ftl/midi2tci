package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"gitlab.com/gomidi/midi"
	"gitlab.com/gomidi/midi/midimessage/channel"
	"gitlab.com/gomidi/midi/reader"
	"gitlab.com/gomidi/midi/writer"
	driver "gitlab.com/gomidi/rtmididrv"
)

var version string = "develop"

var rootCmd = &cobra.Command{
	Use:   "midi2tci",
	Short: "Control ExpertSDR with a MIDI input device through TCI",
	Run:   run,
}

var rootFlags = struct {
	portNumber int
	portName   string
}{}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	log.Printf("midi2tci %s", version)
	rootCmd.PersistentFlags().IntVar(&rootFlags.portNumber, "portNumber", 0, "number of the MIDI port (use list to find out the available ports)")
	rootCmd.PersistentFlags().StringVar(&rootFlags.portName, "portName", "", "name of the MIDI port (use list to find out the available ports)")
}

func run(_ *cobra.Command, _ []string) {
	ctx, done := signal.NotifyContext(context.Background(), os.Interrupt)
	defer done()

	var portNumber int
	var portName string

	if rootFlags.portName != "" {
		portNumber = -1
		portName = rootFlags.portName
	} else {
		portNumber = rootFlags.portNumber
		portName = ""
	}

	drv, err := driver.New()
	if err != nil {
		log.Fatal(err)
	}
	defer drv.Close()

	djControlOut, err := midi.OpenOut(drv, portNumber, portName)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Opened %s successfully for writing", djControlOut)
	wr := writer.New(djControlOut)

	djControlIn, err := midi.OpenIn(drv, portNumber, portName)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Opened %s successfully for reading", djControlIn)
	rd := reader.New(
		reader.NoLogger(),
		reader.Each(func(_ *reader.Position, msg midi.Message) {
			log.Printf("rx: %#v", msg)
			switch m := msg.(type) {
			case channel.NoteOn:
				if (m.Key() != 0x0f && m.Key() != 0x10) || m.Channel() == 0x06 || m.Channel() == 0x07 {
					wr.Write(m)
				}
			case channel.NoteOff:
				if (m.Key() != 0x0f && m.Key() != 0x10) || m.Channel() == 0x06 || m.Channel() == 0x07 {
					wr.Write(m)
				}
			}
		}),
	)
	err = rd.ListenTo(djControlIn)
	if err != nil {
		log.Fatal(err)
	}

	<-ctx.Done()
}
