package ctrl

import (
	"fmt"
	"log"

	"github.com/ftl/tci/client"
)

const (
	EnableRXMapping     MappingType = "enable_rx"
	RXVolumeMapping     MappingType = "rx_volume"
	SetRXVolumeMapping  MappingType = "set_rx_volume"
	RXBalanceMapping    MappingType = "rx_balance"
	SetRXBalanceMapping MappingType = "set_rx_balance"
)

func init() {
	Factories[EnableRXMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControlType, error) {
		vfo, err := AtoVFO(m.VFO)
		if err != nil {
			return nil, 0, err
		}
		return NewRXChannelEnableButton(m.MidiKey(), m.TRX, vfo, led, tciClient), ButtonControl, nil
	}
	Factories[RXVolumeMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControlType, error) {
		vfo, err := AtoVFO(m.VFO)
		if err != nil {
			return nil, 0, err
		}
		controlType, stepSize, reverseDirection, dynamicMode, err := m.ValueControlOptions(1)
		if err != nil {
			return nil, 0, err
		}
		return NewRXVolumeControl(m.MidiKey(), m.TRX, vfo, controlType, led, stepSize, reverseDirection, dynamicMode, tciClient), controlType, nil
	}
	Factories[SetRXVolumeMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControlType, error) {
		vfo, err := AtoVFO(m.VFO)
		if err != nil {
			return nil, 0, err
		}
		value, set, err := m.RequiredIntOption("volume")
		if err != nil {
			return nil, 0, err
		}
		if !set {
			return nil, 0, fmt.Errorf("no volume configured. Use options[\"volume\"]=\"<volume in dB (-60 to 0)>\" to configure the volume")
		}
		return NewSetRXVolumeButton(m.MidiKey(), m.TRX, vfo, led, value, tciClient), ButtonControl, nil
	}
	Factories[RXBalanceMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControlType, error) {
		vfo, err := AtoVFO(m.VFO)
		if err != nil {
			return nil, 0, err
		}
		controlType, stepSize, reverseDirection, dynamicMode, err := m.ValueControlOptions(1)
		if err != nil {
			return nil, 0, err
		}
		return NewRXBalanceControl(m.MidiKey(), m.TRX, vfo, controlType, led, stepSize, reverseDirection, dynamicMode, tciClient), controlType, nil
	}
	Factories[SetRXBalanceMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControlType, error) {
		vfo, err := AtoVFO(m.VFO)
		if err != nil {
			return nil, 0, err
		}
		value, set, err := m.RequiredIntOption("balance")
		if err != nil {
			return nil, 0, err
		}
		if !set {
			return nil, 0, fmt.Errorf("no balance configured. Use options[\"balance\"]=\"<balance in the range -40 to 40>\" to configure the balance")
		}
		return NewSetRXBalanceButton(m.MidiKey(), m.TRX, vfo, led, value, tciClient), ButtonControl, nil
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
	b.led.SetOn(b.key, enabled)
}

func NewRXVolumeControl(key MidiKey, trx int, vfo client.VFO, controlType ControlType, led LED, stepSize int, reverseDirection bool, dynamicMode bool, controller RXVolumeController) *RXVolumeControl {
	set := func(v int) {
		err := controller.SetRXVolume(trx, vfo, v)
		if err != nil {
			log.Printf("Cannot change RX volume: %v", err)
		}
	}
	valueRange := StaticRange{-60, 0}

	return &RXVolumeControl{
		ValueControl: NewValueControl(key, controlType, set, valueRange, led, stepSize, reverseDirection, dynamicMode),
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

func NewSetRXVolumeButton(key MidiKey, trx int, vfo client.VFO, led LED, value int, controller RXVolumeController) *SetRXVolumeButton {
	return &SetRXVolumeButton{
		key:        key,
		trx:        trx,
		vfo:        vfo,
		led:        led,
		controller: controller,
		value:      value,
	}
}

type SetRXVolumeButton struct {
	key        MidiKey
	trx        int
	vfo        client.VFO
	led        LED
	controller RXVolumeController

	value int
}

func (b *SetRXVolumeButton) Pressed() {
	err := b.controller.SetRXVolume(b.trx, b.vfo, b.value)
	if err != nil {
		log.Print(err)
	}
}

func (b *SetRXVolumeButton) SetRXVolume(trx int, vfo client.VFO, volume int) {
	if trx != b.trx || vfo != b.vfo {
		return
	}
	b.led.SetOn(b.key, volume == b.value)
}

func NewRXBalanceControl(key MidiKey, trx int, vfo client.VFO, controlType ControlType, led LED, stepSize int, reverseDirection bool, dynamicMode bool, controller RXBalanceController) *RXBalanceControl {
	set := func(v int) {
		err := controller.SetRXBalance(trx, vfo, v)
		if err != nil {
			log.Printf("Cannot change RX balance: %v", err)
		}
	}
	valueRange := StaticRange{-40, 40}

	return &RXBalanceControl{
		ValueControl: NewValueControl(key, controlType, set, valueRange, led, stepSize, reverseDirection, dynamicMode),
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

func NewSetRXBalanceButton(key MidiKey, trx int, vfo client.VFO, led LED, value int, controller RXBalanceController) *SetRXBalanceButton {
	return &SetRXBalanceButton{
		key:        key,
		trx:        trx,
		vfo:        vfo,
		led:        led,
		controller: controller,
		value:      value,
	}
}

type SetRXBalanceButton struct {
	key        MidiKey
	trx        int
	vfo        client.VFO
	led        LED
	controller RXBalanceController

	value int
}

func (b *SetRXBalanceButton) Pressed() {
	err := b.controller.SetRXBalance(b.trx, b.vfo, b.value)
	if err != nil {
		log.Print(err)
	}
}

func (b *SetRXBalanceButton) SetRXBalance(trx int, vfo client.VFO, balance int) {
	if trx != b.trx || vfo != b.vfo {
		return
	}
	b.led.SetOn(b.key, balance == b.value)
}
