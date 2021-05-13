package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/ftl/tci/client"
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
	Short: "Control ExpertSDR through TCI with a MIDI input device",
	Run:   run,
}

var rootFlags = struct {
	portNumber int
	portName   string
	tciAddress string
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
	rootCmd.PersistentFlags().StringVar(&rootFlags.tciAddress, "tci", "localhost:40001", "the address of the TCI server")
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

	tciHost, err := parseTCIAddr(rootFlags.tciAddress)
	if err != nil {
		log.Fatal(err)
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

	tciHandler := NewTCIHandler(wr)
	tciClient := client.KeepOpen(tciHost, 10*time.Second, tciHandler)
	tciClient.SetTimeout(1 * time.Second)

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
				if m.Channel() == 0x01 && m.Key() == 0x0c {
					muted, err := tciClient.Mute()
					if err != nil {
						log.Print(err)
						break
					}
					tciClient.SetMute(!muted)
				}
				// if (m.Key() != 0x0f && m.Key() != 0x10) || m.Channel() == 0x06 || m.Channel() == 0x07 {
				// 	wr.Write(m)
				// }
			case channel.NoteOff:
			// if (m.Key() != 0x0f && m.Key() != 0x10) || m.Channel() == 0x06 || m.Channel() == 0x07 {
			// 	wr.Write(m)
			// }
			case channel.ControlChange:
				if m.Controller() == 0x0a {
					var vfo client.VFO
					if m.Channel() == 0x01 {
						vfo = client.VFOA
					} else if m.Channel() == 0x02 {
						vfo = client.VFOB
					}
					frequency, err := tciClient.VFOFrequency(0, vfo)
					if err != nil {
						log.Print(err)
						break
					}
					var delta int
					if m.Value() < 0x40 {
						delta = 1 // * velocity
					} else {
						delta = -1 // * velocity
					}
					tciClient.SetVFOFrequency(0, vfo, frequency+delta)
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

func parseTCIAddr(arg string) (*net.TCPAddr, error) {
	host, port := splitHostPort(arg)
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = strconv.Itoa(client.DefaultPort)
	}

	return net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%s", host, port))
}

func splitHostPort(hostport string) (host, port string) {
	host = hostport

	colon := strings.LastIndexByte(host, ':')
	if colon != -1 && validOptionalPort(host[colon:]) {
		host, port = host[:colon], host[colon+1:]
	}

	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}

	return
}

func validOptionalPort(port string) bool {
	if port == "" {
		return true
	}
	if port[0] != ':' {
		return false
	}
	for _, b := range port[1:] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

func NewTCIHandler(w writer.ChannelWriter) *TCIHandler {
	return &TCIHandler{
		w: w,
	}
}

type TCIHandler struct {
	w writer.ChannelWriter
}

func (h *TCIHandler) SetMute(muted bool) {
	if muted {
		h.w.SetChannel(1)
		writer.NoteOff(h.w, 0x0c)
	} else {
		h.w.SetChannel(1)
		writer.NoteOn(h.w, 0x0c, 0x7f)
	}
}
