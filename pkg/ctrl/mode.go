package ctrl

import (
	"fmt"
	"log"
	"strings"

	"github.com/ftl/tci/client"
)

const ModeMapping MappingType = "mode"

func init() {
	Factories[ModeMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControllerType, error) {
		mode, ok := m.Options["mode"]
		if !ok {
			return nil, ButtonController, fmt.Errorf("No mode configured. Use options[\"mode\"]=\"<mode>\" to configure the mode you want to select.")

		}
		mode = strings.TrimSpace(strings.ToLower(mode))
		return NewModeButton(m.MidiKey(), m.TRX, client.Mode(mode), led, tciClient), ButtonController, nil
	}
}

func NewModeButton(key MidiKey, trx int, mode client.Mode, led LED, controller ModeController) *ModeButton {
	return &ModeButton{
		key:        key,
		trx:        trx,
		led:        led,
		controller: controller,

		mode: mode,
	}
}

type ModeButton struct {
	key        MidiKey
	trx        int
	led        LED
	controller ModeController

	mode client.Mode

	enabled bool
}

type ModeController interface {
	SetMode(trx int, mode client.Mode) error
}

func (b *ModeButton) Pressed() {
	err := b.controller.SetMode(b.trx, b.mode)
	if err != nil {
		log.Print(err)
	}
}

func (b *ModeButton) SetMode(trx int, mode client.Mode) {
	if trx != b.trx {
		return
	}
	b.enabled = (mode == b.mode)
	b.led.Set(b.key, b.enabled)
}
