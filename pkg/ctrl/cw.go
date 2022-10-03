package ctrl

import (
	"log"

	"github.com/ftl/tci/client"
)

const (
	SendCWMapping  MappingType = "send_cw"
	StopCWMapping  MappingType = "stop_cw"
	CWSpeedMapping MappingType = "cw_speed"
)

func init() {
	Factories[SendCWMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControlType, error) {
		text := m.Options["text"]
		return NewSendCWButton(m.MidiKey(), m.TRX, led, text, tciClient), ButtonControl, nil
	}
	Factories[StopCWMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControlType, error) {
		return NewStopCWButton(m.MidiKey(), m.TRX, led, tciClient), ButtonControl, nil
	}
	Factories[CWSpeedMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControlType, error) {
		controlType, stepSize, reverseDirection, dynamicMode, err := m.ValueControlOptions(1)
		if err != nil {
			return nil, 0, err
		}
		return NewCWSpeedControl(m.MidiKey(), controlType, led, stepSize, reverseDirection, dynamicMode, tciClient), controlType, nil
	}
}

func NewSendCWButton(key MidiKey, trx int, led LED, text string, controller CWController) *SendCWButton {
	return &SendCWButton{
		key:        key,
		trx:        trx,
		led:        led,
		text:       text,
		controller: controller,
	}
}

type SendCWButton struct {
	key        MidiKey
	trx        int
	led        LED
	text       string
	controller CWController
}

type CWController interface {
	SendCWMessage(int, string, string, string) error
	StopCW() error
	SetCWMacrosSpeed(wpm int) error
}

func (b *SendCWButton) Pressed() {
	err := b.controller.SendCWMessage(b.trx, b.text, "", "")
	if err != nil {
		log.Print(err)
	}
}

func NewStopCWButton(key MidiKey, trx int, led LED, controller CWController) *StopCWButton {
	return &StopCWButton{
		key:        key,
		trx:        trx,
		led:        led,
		controller: controller,
	}
}

type StopCWButton struct {
	key        MidiKey
	trx        int
	led        LED
	controller CWController
}

func (b *StopCWButton) Pressed() {
	err := b.controller.StopCW()
	if err != nil {
		log.Print(err)
	}
}

func NewCWSpeedControl(key MidiKey, controlType ControlType, led LED, stepSize int, reverseDirection bool, dynamicMode bool, controller CWController) *CWSpeedControl {
	set := func(v int) {
		err := controller.SetCWMacrosSpeed(v)
		if err != nil {
			log.Printf("Cannot change RX balance: %v", err)
		}
	}
	valueRange := StaticRange{5, 50}

	return &CWSpeedControl{
		ValueControl: NewValueControl(key, controlType, set, valueRange, led, stepSize, reverseDirection, dynamicMode),
	}
}

type CWSpeedControl struct {
	ValueControl
}

type CWSpeedController interface {
	SetRXBalance(trx int, vfo client.VFO, dB int) error
}

func (s *CWSpeedControl) SetCWMacrosSpeed(wpm int) {
	s.ValueControl.SetActiveValue(wpm)
}
