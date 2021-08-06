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
	rootCmd.PersistentFlags().BoolVar(&rootFlags.trace, "trace", false, "print a trace of all incoming MIDI messages")
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

	config := cfg.Configuration{
		PortNumber: portNumber,
		PortName:   portName,
		Mappings: []ctrl.Mapping{
			{Type: ctrl.VFOMapping, Channel: 1, Key: 0x0a, TRX: 0, VFO: "VFOA"},
			{Type: ctrl.VFOMapping, Channel: 2, Key: 0x0a, TRX: 0, VFO: "VFOB"},
			{Type: ctrl.MuteMapping, Channel: 1, Key: 0x0c},
			{Type: ctrl.VolumeMapping, Channel: 0, Key: 0x03},
			{Type: ctrl.ModeMapping, Channel: 7, Key: 0x00, TRX: 0, Options: map[string]string{"mode": "CW"}},
			{Type: ctrl.ModeMapping, Channel: 7, Key: 0x01, TRX: 0, Options: map[string]string{"mode": "DIGU"}},
			{Type: ctrl.ModeMapping, Channel: 7, Key: 0x02, TRX: 0, Options: map[string]string{"mode": "LSB"}},
			{Type: ctrl.ModeMapping, Channel: 7, Key: 0x03, TRX: 0, Options: map[string]string{"mode": "USB"}},
			{Type: ctrl.FilterMapping, Channel: 6, Key: 0x02, TRX: 0, Options: map[string]string{"min": "-50", "max": "50"}},
			{Type: ctrl.FilterMapping, Channel: 6, Key: 0x03, TRX: 0, Options: map[string]string{"min": "1400", "max": "1600"}},
		},
	}

	// setup the outgoing MIDI communication
	djControlOut, err := midi.OpenOut(drv, config.PortNumber, config.PortName)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Opened %s successfully for writing", djControlOut)
	wr := writer.New(djControlOut)
	ledController := NewLEDController(wr)
	defer ledController.Close()

	// open the TCI connection
	tciClient := client.KeepOpen(tciHost, 10*time.Second)
	tciClient.SetTimeout(500 * time.Millisecond)

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

	// muteKey := ctrl.MidiKey{Channel: 1, Key: 0x0c}
	// muteButton := ctrl.NewMuteButton(muteKey, ledController, tciClient)
	// buttons[muteKey] = muteButton
	// tciClient.Notify(muteButton)

	// rxChannelEnableKey := ctrl.MidiKey{Channel: 2, Key: 0x0c}
	// rxChannelEnableButton := ctrl.NewRXChannelEnableButton(rxChannelEnableKey, 0, client.VFOB, ledController, tciClient)
	// buttons[rxChannelEnableKey] = rxChannelEnableButton
	// tciClient.Notify(rxChannelEnableButton)

	// splitEnableKey := ctrl.MidiKey{Channel: 2, Key: 0x03}
	// splitEnableButton := ctrl.NewSplitEnableButton(splitEnableKey, 0, ledController, tciClient)
	// buttons[splitEnableKey] = splitEnableButton
	// tciClient.Notify(splitEnableButton)

	// syncAWithBKey := ctrl.MidiKey{Channel: 1, Key: 0x05}
	// syncAWithBButton := ctrl.NewSyncVFOFrequencyButton(0, client.VFOB, 0, client.VFOA, tciClient, tciClient)
	// buttons[syncAWithBKey] = syncAWithBButton

	// syncBWithAKey := ctrl.MidiKey{Channel: 2, Key: 0x05}
	// syncBWithAButton := ctrl.NewSyncVFOFrequencyButton(0, client.VFOA, 0, client.VFOB, tciClient, tciClient)
	// buttons[syncBWithAKey] = syncBWithAButton

	// vfo1Key := ctrl.MidiKey{Channel: 1, Key: 0x0a}
	// vfo1Wheel := ctrl.NewVFOWheel(vfo1Key, 0, client.VFOA, tciClient)
	// defer vfo1Wheel.Close()
	// wheels[vfo1Key] = vfo1Wheel
	// tciClient.Notify(vfo1Wheel)

	// vfo2Key := ctrl.MidiKey{Channel: 2, Key: 0x0a}
	// vfo2Wheel := ctrl.NewVFOWheel(vfo2Key, 0, client.VFOB, tciClient)
	// defer vfo2Wheel.Close()
	// wheels[vfo2Key] = vfo2Wheel
	// tciClient.Notify(vfo2Wheel)

	// volumeKey := ctrl.MidiKey{Channel: 0, Key: 0x03}
	// volumeSlider := ctrl.NewVolumeSlider(tciClient)
	// defer volumeSlider.Close()
	// sliders[volumeKey] = volumeSlider
	// tciClient.Notify(volumeSlider)

	// vfo1VolumeKey := ctrl.MidiKey{Channel: 1, Key: 0x00}
	// vfo1VolumeSlider := ctrl.NewRXVolumeSlider(0, client.VFOA, tciClient)
	// defer vfo1VolumeSlider.Close()
	// sliders[vfo1VolumeKey] = vfo1VolumeSlider
	// tciClient.Notify(vfo1VolumeSlider)

	// vfo2VolumeKey := ctrl.MidiKey{Channel: 2, Key: 0x00}
	// vfo2VolumeSlider := ctrl.NewRXVolumeSlider(0, client.VFOB, tciClient)
	// defer vfo2VolumeSlider.Close()
	// sliders[vfo2VolumeKey] = vfo2VolumeSlider
	// tciClient.Notify(vfo2VolumeSlider)

	// vfo1BalanceKey := ctrl.MidiKey{Channel: 1, Key: 0x02}
	// vfo1BalanceSlider := ctrl.NewRXBalanceSlider(0, client.VFOA, tciClient)
	// defer vfo1BalanceSlider.Close()
	// sliders[vfo1BalanceKey] = vfo1BalanceSlider
	// tciClient.Notify(vfo1BalanceSlider)

	// vfo2BalanceKey := ctrl.MidiKey{Channel: 2, Key: 0x02}
	// vfo2BalanceSlider := ctrl.NewRXBalanceSlider(0, client.VFOB, tciClient)
	// defer vfo2BalanceSlider.Close()
	// sliders[vfo2BalanceKey] = vfo2BalanceSlider
	// tciClient.Notify(vfo2BalanceSlider)

	// rxMixerKey := ctrl.MidiKey{Channel: 0, Key: 0x00}
	// rxMixerSlider := ctrl.NewRXMixer(0, tciClient)
	// defer rxMixerSlider.Close()
	// sliders[rxMixerKey] = rxMixerSlider
	// tciClient.Notify(rxMixerSlider)

	// ritKey := ctrl.MidiKey{Channel: 1, Key: 0x08}
	// ritSlider := ctrl.NewRITSlider(0, tciClient)
	// defer ritSlider.Close()
	// sliders[ritKey] = ritSlider
	// tciClient.Notify(ritSlider)

	// xitKey := ctrl.MidiKey{Channel: 2, Key: 0x08}
	// xitSlider := ctrl.NewXITSlider(0, tciClient)
	// defer xitSlider.Close()
	// sliders[xitKey] = xitSlider
	// tciClient.Notify(xitSlider)

	// ritEnableKey := ctrl.MidiKey{Channel: 6, Key: 0x00}
	// ritEnableButton := ctrl.NewRITEnableButton(ritEnableKey, 0, ledController, tciClient)
	// buttons[ritEnableKey] = ritEnableButton
	// tciClient.Notify(ritEnableButton)

	// xitEnableKey := ctrl.MidiKey{Channel: 6, Key: 0x01}
	// xitEnableButton := ctrl.NewXITEnableButton(xitEnableKey, 0, ledController, tciClient)
	// buttons[xitEnableKey] = xitEnableButton
	// tciClient.Notify(xitEnableButton)

	// cwFilterKey := ctrl.MidiKey{Channel: 6, Key: 0x02}
	// cwFilterButton := ctrl.NewFilterBandButton(cwFilterKey, 0, -50, 50, ledController, tciClient)
	// buttons[cwFilterKey] = cwFilterButton
	// tciClient.Notify(cwFilterButton)

	// rttyFilterKey := ctrl.MidiKey{Channel: 6, Key: 0x03}
	// rttyFilterButton := ctrl.NewFilterBandButton(rttyFilterKey, 0, 1200, 1800, ledController, tciClient)
	// buttons[rttyFilterKey] = rttyFilterButton
	// tciClient.Notify(rttyFilterButton)

	// cwModeKey := ctrl.MidiKey{Channel: 7, Key: 0x00}
	// cwModeButton := ctrl.NewModeButton(cwModeKey, 0, client.ModeCW, ledController, tciClient)
	// buttons[cwModeKey] = cwModeButton
	// tciClient.Notify(cwModeButton)

	// digitalModeKey := ctrl.MidiKey{Channel: 7, Key: 0x01}
	// digitalModeButton := ctrl.NewModeButton(digitalModeKey, 0, client.ModeDIGU, ledController, tciClient)
	// buttons[digitalModeKey] = digitalModeButton
	// tciClient.Notify(digitalModeButton)

	// lsbModeKey := ctrl.MidiKey{Channel: 7, Key: 0x02}
	// lsbModeButton := ctrl.NewModeButton(lsbModeKey, 0, client.ModeLSB, ledController, tciClient)
	// buttons[lsbModeKey] = lsbModeButton
	// tciClient.Notify(lsbModeButton)

	// usbModeKey := ctrl.MidiKey{Channel: 7, Key: 0x03}
	// usbModeButton := ctrl.NewModeButton(usbModeKey, 0, client.ModeUSB, ledController, tciClient)
	// buttons[usbModeKey] = usbModeButton
	// tciClient.Notify(usbModeButton)

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
				var delta int
				if value < 0x40 {
					delta = 1
				} else {
					delta = -1
				}
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
