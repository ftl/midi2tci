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

func init() {
	Factories[EnableRXMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControlType, error) {
		vfo, err := AtoVFO(m.VFO)
		if err != nil {
			return nil, 0, err
		}
		return NewRXChannelEnableButton(m.MidiKey(), m.TRX, vfo, led, tciClient), ButtonControl, nil
	}
	Factories[RXVolumeMapping] = func(m Mapping, _ LED, tciClient *client.Client) (interface{}, ControlType, error) {
		vfo, err := AtoVFO(m.VFO)
		if err != nil {
			return nil, 0, err
		}
		controlType, stepSize, reverseDirection, dynamicMode, err := m.ValueControlOptions(1)
		if err != nil {
			return nil, 0, err
		}
		return NewRXVolumeControl(m.TRX, vfo, controlType, stepSize, reverseDirection, dynamicMode, tciClient), controlType, nil
	}
	Factories[RXBalanceMapping] = func(m Mapping, _ LED, tciClient *client.Client) (interface{}, ControlType, error) {
		vfo, err := AtoVFO(m.VFO)
		if err != nil {
			return nil, 0, err
		}
		controlType, stepSize, reverseDirection, dynamicMode, err := m.ValueControlOptions(1)
		if err != nil {
			return nil, 0, err
		}
		return NewRXBalanceControl(m.TRX, vfo, controlType, stepSize, reverseDirection, dynamicMode, tciClient), controlType, nil
	}
}

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

func NewRXVolumeControl(trx int, vfo client.VFO, controlType ControlType, stepSize int, reverseDirection bool, dynamicMode bool, controller RXVolumeController) *RXVolumeControl {
	set := func(v int) {
		err := controller.SetRXVolume(trx, vfo, v)
		if err != nil {
			log.Printf("Cannot change RX volume: %v", err)
		}
	}
	valueRange := StaticRange{-60, 0}

	return &RXVolumeControl{
		ValueControl: NewValueControl(controlType, set, valueRange, stepSize, reverseDirection, dynamicMode),
		trx:          trx,
		vfo:          vfo,
	}
}

type RXVolumeControl struct {
	ValueControl
	trx int
	vfo client.VFO
}

type RXVolumeController interface {
	SetRXVolume(trx int, vfo client.VFO, dB int) error
}

func (s *RXVolumeControl) SetRXVolume(trx int, vfo client.VFO, volume int) {
	if trx != s.trx || vfo != s.vfo {
		return
	}
	s.ValueControl.SetActiveValue(volume)
}

func NewRXBalanceControl(trx int, vfo client.VFO, controlType ControlType, stepSize int, reverseDirection bool, dynamicMode bool, controller RXBalanceController) *RXBalanceControl {
	set := func(v int) {
		err := controller.SetRXBalance(trx, vfo, v)
		if err != nil {
			log.Printf("Cannot change RX balance: %v", err)
		}
	}
	valueRange := StaticRange{-40, 40}

	return &RXBalanceControl{
		ValueControl: NewValueControl(controlType, set, valueRange, stepSize, reverseDirection, dynamicMode),
		trx:          trx,
		vfo:          vfo,
	}
}

type RXBalanceControl struct {
	ValueControl
	trx int
	vfo client.VFO
}

type RXBalanceController interface {
	SetRXBalance(trx int, vfo client.VFO, dB int) error
}

func (s *RXBalanceControl) SetRXBalance(trx int, vfo client.VFO, balance int) {
	if trx != s.trx || vfo != s.vfo {
		return
	}
	s.ValueControl.SetActiveValue(balance)
}
