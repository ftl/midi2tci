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

	"github.com/ftl/midi2tci/pkg/ctrl"
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
	ledController := NewLEDController(wr)

	// open the TCI connection
	tciClient := client.KeepOpen(tciHost, 10*time.Second)
	tciClient.SetTimeout(1 * time.Second)

	// setup the configured controls
	muteKey := ctrl.MidiKey{Channel: 1, Key: 0x0c}
	muteButton := ctrl.NewMuteButton(muteKey, ledController, tciClient)
	buttons[muteKey] = muteButton
	tciClient.Notify(muteButton)

	vfo1Key := ctrl.MidiKey{Channel: 1, Key: 0x0a}
	vfo1Wheel := ctrl.NewVFOWheel(vfo1Key, 0, client.VFOA, tciClient)
	wheels[vfo1Key] = vfo1Wheel
	tciClient.Notify(vfo1Wheel)

	vfo2Key := ctrl.MidiKey{Channel: 2, Key: 0x0a}
	vfo2Wheel := ctrl.NewVFOWheel(vfo2Key, 0, client.VFOB, tciClient)
	wheels[vfo2Key] = vfo2Wheel
	tciClient.Notify(vfo2Wheel)

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
				key := ctrl.MidiKey{Channel: m.Channel(), Key: m.Key()}
				button, ok := buttons[key]
				if ok {
					button.Pressed()
				}
			case channel.ControlChange:
				key := ctrl.MidiKey{Channel: m.Channel(), Key: m.Controller()}
				wheel, ok := wheels[key]
				if ok {
					var delta int
					if m.Value() < 0x40 {
						delta = 1
					} else {
						delta = -1
					}
					wheel.Turned(delta)
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

func NewLEDController(w writer.ChannelWriter) *LEDController {
	return &LEDController{
		w: w,
	}
}

type LEDController struct {
	w writer.ChannelWriter
}

func (c *LEDController) Set(key ctrl.MidiKey, on bool) {
	c.w.SetChannel(key.Channel)
	if on {
		writer.NoteOn(c.w, key.Key, 0x7f)
	} else {
		writer.NoteOff(c.w, key.Key)
	}
}

type Button interface {
	Pressed()
}

var buttons map[ctrl.MidiKey]Button = make(map[ctrl.MidiKey]Button)

type Wheel interface {
	Turned(delta int)
}

var wheels map[ctrl.MidiKey]Wheel = make(map[ctrl.MidiKey]Wheel)
