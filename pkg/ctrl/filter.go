package ctrl

import "log"

func NewFilterBandButton(key MidiKey, trx int, bottomFrequency int, topFrequency int, led LED, controller RXFilterBandController) *FilterBandButton {
	return &FilterBandButton{
		key:        key,
		trx:        trx,
		led:        led,
		controller: controller,

		bottomFrequency: bottomFrequency,
		topFrequency:    topFrequency,
	}
}

type FilterBandButton struct {
	key        MidiKey
	trx        int
	led        LED
	controller RXFilterBandController

	bottomFrequency int
	topFrequency    int

	enabled bool
}

type RXFilterBandController interface {
	SetRXFilterBand(trx int, min, max int) error
}

func (b *FilterBandButton) Pressed() {
	err := b.controller.SetRXFilterBand(b.trx, b.bottomFrequency, b.topFrequency)
	if err != nil {
		log.Print(err)
	}
}

func (b *FilterBandButton) SetRXFilterBand(trx int, min, max int) {
	if trx != b.trx {
		return
	}
	b.enabled = (min == b.bottomFrequency) && (max == b.topFrequency)
	b.led.Set(b.key, b.enabled)
}
