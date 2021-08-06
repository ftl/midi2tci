package ctrl

import (
	"fmt"
	"log"
	"strconv"

	"github.com/ftl/tci/client"
)

const (
	FilterMapping MappingType = "filter"
)

func init() {
	Factories[FilterMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControllerType, error) {
		minFrequencyStr, ok := m.Options["min"]
		if !ok {
			return nil, ButtonController, fmt.Errorf("No minimum frequency configured. Use options[\"min\"]=\"<min frequency in Hz>\" to configure the filter's minimum frequency.")
		}
		minFrequency, err := strconv.Atoi(minFrequencyStr)
		if err != nil {
			return nil, ButtonController, fmt.Errorf("Invalid minimum frequency %s: %v", minFrequencyStr, err)
		}

		maxFrequencyStr, ok := m.Options["max"]
		if !ok {
			return nil, ButtonController, fmt.Errorf("No maximum frequency configured. Use options[\"max\"]=\"<max frequency in Hz>\" to configure the filter's maximum frequency.")
		}
		maxFrequency, err := strconv.Atoi(maxFrequencyStr)
		if err != nil {
			return nil, ButtonController, fmt.Errorf("Invalid maximum frequency %s: %v", maxFrequencyStr, err)
		}

		return NewFilterBandButton(m.MidiKey(), m.TRX, minFrequency, maxFrequency, led, tciClient), ButtonController, nil
	}
}

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
