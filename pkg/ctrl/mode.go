package ctrl

import (
	"log"

	"github.com/ftl/tci/client"
)

const ModeMapping MappingType = "mode"

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
