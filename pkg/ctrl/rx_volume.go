package ctrl

import (
	"log"

	"github.com/ftl/tci/client"
)

const (
	EnableRXMapping  MappingType = "enable_rx"
	RXVolumeMapping  MappingType = "rx_volume"
	RXBalanceMapping MappingType = "rx_balance"
)

func NewRXChannelEnableButton(key MidiKey, trx int, vfo client.VFO, led LED, rxChannelEnabler RXChannelEnabler) *RXChannelEnableButton {
	return &RXChannelEnableButton{
		key:              key,
		trx:              trx,
		vfo:              vfo,
		led:              led,
		rxChannelEnabler: rxChannelEnabler,
	}
}

type RXChannelEnableButton struct {
	key              MidiKey
	trx              int
	vfo              client.VFO
	led              LED
	rxChannelEnabler RXChannelEnabler

	enabled bool
}

type RXChannelEnabler interface {
	SetRXChannelEnable(int, client.VFO, bool) error
}

func (b *RXChannelEnableButton) Pressed() {
	err := b.rxChannelEnabler.SetRXChannelEnable(b.trx, b.vfo, !b.enabled)
	if err != nil {
		log.Print(err)
	}
}

func (b *RXChannelEnableButton) SetRXChannelEnable(trx int, vfo client.VFO, enabled bool) {
	if trx != b.trx || vfo != b.vfo {
		return
	}
	b.enabled = enabled
	b.led.Set(b.key, enabled)
}

func NewRXVolumeSlider(trx int, vfo client.VFO, controller RXVolumeController) *RXVolumeSlider {
	const tick = float64(60.0 / 127.0)
	return &RXVolumeSlider{
		Slider: NewSlider(
			func(v int) {
				err := controller.SetRXVolume(trx, vfo, v)
				if err != nil {
					log.Printf("Cannot change RX volume: %v", err)
				}
			},
			func(v int) int { return -60 + int(float64(v)*tick) },
		),
		trx: trx,
		vfo: vfo,
	}
}

type RXVolumeSlider struct {
	*Slider
	trx int
	vfo client.VFO
}

type RXVolumeController interface {
	SetRXVolume(trx int, vfo client.VFO, dB int) error
}

func (s *RXVolumeSlider) SetRXVolume(trx int, vfo client.VFO, volume int) {
	if trx != s.trx || vfo != s.vfo {
		return
	}
	s.Slider.SetActiveValue(volume)
}

func NewRXBalanceSlider(trx int, vfo client.VFO, controller RXBalanceController) *RXBalanceSlider {
	const tick = float64(80.0 / 127.0)
	return &RXBalanceSlider{
		Slider: NewSlider(
			func(v int) {
				err := controller.SetRXBalance(trx, vfo, v)
				if err != nil {
					log.Printf("Cannot change RX balance: %v", err)
				}
			},
			func(v int) int { return -40 + int(float64(v)*tick) },
		),
		trx: trx,
		vfo: vfo,
	}
}

type RXBalanceSlider struct {
	*Slider
	trx int
	vfo client.VFO
}

type RXBalanceController interface {
	SetRXBalance(trx int, vfo client.VFO, dB int) error
}

func (s *RXBalanceSlider) SetRXBalance(trx int, vfo client.VFO, balance int) {
	if trx != s.trx || vfo != s.vfo {
		return
	}
	s.Slider.SetActiveValue(balance)
}
