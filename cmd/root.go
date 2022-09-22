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
	"gitlab.com/gomidi/midi/reader"
	"gitlab.com/gomidi/midi/writer"
	driver "gitlab.com/gomidi/rtmididrv"

	"github.com/ftl/midi2tci/pkg/cfg"
	"github.com/ftl/midi2tci/pkg/ctrl"
)

var version string = "develop"

var rootCmd = &cobra.Command{
	Use:   "midi2tci",
	Short: "Control ExpertSDR through TCI with a MIDI input device",
	Run:   run,
}

var rootFlags = struct {
	trace      bool
	traceTci   bool
	portNumber int
	portName   string
	tciAddress string
	configFile string
}{}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	log.Printf("midi2tci %s", version)
	rootCmd.PersistentFlags().IntVar(&rootFlags.portNumber, "portNumber", -1, "number of the MIDI port (use list to find out the available ports)")
	rootCmd.PersistentFlags().StringVar(&rootFlags.portName, "portName", "", "name of the MIDI port (use list to find out the available ports)")
	rootCmd.PersistentFlags().StringVar(&rootFlags.tciAddress, "tci", "", "the address of the TCI server")
	rootCmd.PersistentFlags().BoolVar(&rootFlags.trace, "trace", false, "print a trace of all incoming MIDI messages")
	rootCmd.PersistentFlags().BoolVar(&rootFlags.traceTci, "traceTci", false, "print tracing information of the TCI client")
	rootCmd.PersistentFlags().StringVar(&rootFlags.configFile, "config", "./config.json", "the configuration file")
}

func run(_ *cobra.Command, _ []string) {
	ctx, done := signal.NotifyContext(context.Background(), os.Interrupt)
	defer done()

	config, err := cfg.ReadFile(rootFlags.configFile)
	if err != nil {
		log.Printf("Cannot read configuration file: %v", err)
		config = cfg.Configuration{}
	} else {
		log.Printf("Using configuration from %s", rootFlags.configFile)
	}

	var portNumber int
	var portName string
	if rootFlags.portName != "" {
		portNumber = -1
		portName = rootFlags.portName
	} else if rootFlags.portNumber >= 0 {
		portNumber = rootFlags.portNumber
		portName = ""
	} else if config.PortName != "" {
		portNumber = -1
		portName = config.PortName
	} else {
		portNumber = config.PortNumber
		portName = ""
	}

	var tciHost *net.TCPAddr
	if rootFlags.tciAddress != "" {
		tciHost, err = parseTCIAddr(rootFlags.tciAddress)
	} else {
		tciHost, err = parseTCIAddr(config.TCIAddress)
	}
	if err != nil {
		log.Fatalf("Invalid TCI host address: %v", err)
	}

	drv, err := driver.New()
	if err != nil {
		log.Fatal(err)
	}
	defer drv.Close()

	// setup the outgoing MIDI communication
	djControlOut, err := midi.OpenOut(drv, portNumber, portName)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Opened %s successfully for writing", djControlOut)
	wr := writer.New(djControlOut)

	if len(config.InitSequence) > 0 {
		log.Print("MIDI init sequence")
		err = SendRawMidiSequence(wr, config.InitSequence)
		if err != nil {
			log.Fatal(err)
		}
	}

	ledController := NewLEDController(wr)
	defer ledController.Close()

	// open the TCI connection
	tciClient := client.KeepOpen(tciHost, 10*time.Second, rootFlags.traceTci)
	tciClient.Notify(&connectionListener{
		midiWriter:         wr,
		connectSequence:    config.ConnectSequence,
		disconnectSequence: config.DisconnectSequence,
	})

	// setup the configured controls
	for _, mapping := range config.Mappings {
		newController, ok := ctrl.Factories[mapping.Type]
		if !ok {
			log.Printf("Cannot find factory for %s", mapping.Type)
			continue
		}

		controller, controllerType, err := newController(mapping, ledController, tciClient)
		if err != nil {
			log.Printf("Cannot create %s: %v", mapping.Type, err)
			continue
		}

		switch controllerType {
		case ctrl.ButtonController:
			button := controller.(Button)
			buttons[mapping.MidiKey()] = button
		case ctrl.SliderController:
			slider := controller.(Slider)
			defer slider.Close()
			sliders[mapping.MidiKey()] = slider
		case ctrl.WheelController:
			wheel := controller.(Wheel)
			defer wheel.Close()
			wheels[mapping.MidiKey()] = wheel
		}
		tciClient.Notify(controller)
	}

	// setup the incoming MIDI communication
	djControlIn, err := midi.OpenIn(drv, portNumber, portName)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Opened %s successfully for reading", djControlIn)
	rd := reader.New(
		reader.NoLogger(),
		reader.NoteOn(func(_ *reader.Position, channel, key, velocity uint8) {
			button, ok := buttons[ctrl.MidiKey{Channel: channel, Key: key}]
			if ok {
				button.Pressed()
			}
		}),
		reader.ControlChange(func(_ *reader.Position, channel, controller, value uint8) {
			midiKey := ctrl.MidiKey{Channel: channel, Key: controller}
			wheel, ok := wheels[midiKey]
			if ok {
				delta := int(value) - int(0x40)
				wheel.Turned(delta)
			}
			slider, ok := sliders[midiKey]
			if ok {
				slider.Changed(int(value))
			}
		}),
		reader.Each(func(_ *reader.Position, msg midi.Message) {
			if rootFlags.trace {
				log.Printf("rx: %#v", msg)
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

type connectionListener struct {
	midiWriter         writer.ChannelWriter
	connectSequence    [][]byte
	disconnectSequence [][]byte
}

func (l *connectionListener) Connected(connected bool) {
	if connected {
		SendRawMidiSequence(l.midiWriter, l.connectSequence)
	} else {
		SendRawMidiSequence(l.midiWriter, l.disconnectSequence)
	}
}

func SendRawMidiSequence(w writer.ChannelWriter, sequence [][]byte) error {
	messages := make([]midi.Message, len(sequence))
	for i, raw := range sequence {
		messages[i] = NewRawMessage(raw)
		if rootFlags.trace {
			log.Printf("%s", messages[i])
		}
	}
	return writer.WriteMessages(w, messages)
}

func NewRawMessage(raw []byte) midi.Message {
	return rawMessage(raw)
}

type rawMessage []byte

func (m rawMessage) String() string {
	return fmt.Sprintf("raw MIDI message: % 2X", []byte(m))
}

func (m rawMessage) Raw() []byte {
	return []byte(m)
}

func NewLEDController(w writer.ChannelWriter) *LEDController {
	result := &LEDController{
		w:        w,
		commands: make(chan func(writer.ChannelWriter)),
		closed:   make(chan struct{}),
	}

	go func() {
		defer close(result.closed)
		for command := range result.commands {
			command(result.w)
		}
	}()

	return result
}

func (c *LEDController) Close() {
	select {
	case <-c.closed:
		return
	default:
		close(c.commands)
		<-c.closed
	}
}

type LEDController struct {
	w        writer.ChannelWriter
	commands chan func(writer.ChannelWriter)
	closed   chan struct{}
}

func (c *LEDController) Set(key ctrl.MidiKey, on bool) {
	c.commands <- func(w writer.ChannelWriter) {
		w.SetChannel(key.Channel)
		if on {
			writer.NoteOn(w, key.Key, 0x7f)
		} else {
			writer.NoteOff(w, key.Key)
		}
	}
}

type Button interface {
	Pressed()
}

var buttons map[ctrl.MidiKey]Button = make(map[ctrl.MidiKey]Button)

type Wheel interface {
	Turned(delta int)
	Close()
}

var wheels map[ctrl.MidiKey]Wheel = make(map[ctrl.MidiKey]Wheel)

type Slider interface {
	Changed(value int)
	Close()
}

var sliders map[ctrl.MidiKey]Slider = make(map[ctrl.MidiKey]Slider)
