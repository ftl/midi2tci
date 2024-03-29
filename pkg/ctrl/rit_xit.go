package ctrl

import (
	"log"

	"github.com/ftl/tci/client"
)

const (
	defaultRitXitRange = 100

	EnableRITMapping MappingType = "enable_rit"
	RITMapping       MappingType = "rit"
	EnableXITMapping MappingType = "enable_xit"
	XITMapping       MappingType = "xit"
)

func init() {
	Factories[EnableRITMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControlType, error) {
		reset := m.BoolOption("reset", false)
		return NewRITEnableButton(m.MidiKey(), m.TRX, led, reset, tciClient), ButtonControl, nil
	}
	Factories[RITMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControlType, error) {
		controlType, stepSize, reverseDirection, dynamicMode, err := m.ValueControlOptions(1)
		if err != nil {
			return nil, 0, err
		}
		frequencyRange, err := m.IntOption("range", defaultRitXitRange)
		if err != nil {
			return nil, 0, err
		}
		return NewRITControl(m.MidiKey(), m.TRX, controlType, led, stepSize, reverseDirection, dynamicMode, frequencyRange, tciClient), controlType, nil
	}
	Factories[EnableXITMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControlType, error) {
		reset := m.BoolOption("reset", false)
		return NewXITEnableButton(m.MidiKey(), m.TRX, led, reset, tciClient), ButtonControl, nil
	}
	Factories[XITMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControlType, error) {
		controlType, stepSize, reverseDirection, dynamicMode, err := m.ValueControlOptions(1)
		if err != nil {
			return nil, 0, err
		}
		frequencyRange, err := m.IntOption("range", defaultRitXitRange)
		if err != nil {
			return nil, 0, err
		}
		return NewXITControl(m.MidiKey(), m.TRX, controlType, led, stepSize, reverseDirection, dynamicMode, frequencyRange, tciClient), controlType, nil
	}
}

func NewRITEnableButton(key MidiKey, trx int, led LED, reset bool, ritEnabler RITEnabler) *RITEnableButton {
	return &RITEnableButton{
		key:        key,
		trx:        trx,
		led:        led,
		ritEnabler: ritEnabler,
		reset:      reset,
	}
}

type RITEnableButton struct {
	key        MidiKey
	trx        int
	led        LED
	ritEnabler RITEnabler

	reset   bool
	enabled bool
}

type RITEnabler interface {
	SetRITOffset(int, int) error
	SetRITEnable(int, bool) error
}

func (b *RITEnableButton) Pressed() {
	err := b.ritEnabler.SetRITEnable(b.trx, !b.enabled)
	if err != nil {
		log.Print(err)
	}
	if b.reset {
		err := b.ritEnabler.SetRITOffset(b.trx, 0)
		if err != nil {
			log.Print(err)
		}
	}
}

func (b *RITEnableButton) SetRITEnable(trx int, enabled bool) {
	if trx != b.trx {
		return
	}
	b.enabled = enabled
	b.led.SetOn(b.key, enabled)
}

func NewXITEnableButton(key MidiKey, trx int, led LED, reset bool, xitEnabler XITEnabler) *XITEnableButton {
	return &XITEnableButton{
		key:        key,
		trx:        trx,
		led:        led,
		xitEnabler: xitEnabler,
		reset:      reset,
	}
}

type XITEnableButton struct {
	key        MidiKey
	trx        int
	led        LED
	xitEnabler XITEnabler

	reset   bool
	enabled bool
}

type XITEnabler interface {
	SetXITOffset(int, int) error
	SetXITEnable(int, bool) error
}

func (b *XITEnableButton) Pressed() {
	err := b.xitEnabler.SetXITEnable(b.trx, !b.enabled)
	if err != nil {
		log.Print(err)
	}
	if b.reset {
		err := b.xitEnabler.SetXITOffset(b.trx, 0)
		if err != nil {
			log.Print(err)
		}
	}
}

func (b *XITEnableButton) SetXITEnable(trx int, enabled bool) {
	if trx != b.trx {
		return
	}
	b.enabled = enabled
	b.led.SetOn(b.key, enabled)
}

func NewRITControl(key MidiKey, trx int, controlType ControlType, led LED, stepSize int, reverseDirection bool, dynamicMode bool, frequencyRange int, controller RITController) *RITControl {
	set := func(v int) {
		err := controller.SetRITOffset(trx, v)
		if err != nil {
			log.Printf("Cannot change RIT offset: %v", err)
		}
	}
	valueRange := StaticRange{-frequencyRange, frequencyRange}
	return &RITControl{
		ValueControl: NewValueControl(key, controlType, set, valueRange, led, stepSize, reverseDirection, dynamicMode),
		trx:          trx,
	}
}

type RITControl struct {
	ValueControl
	trx int
}

type RITController interface {
	SetRITOffset(trx int, offset int) error
}

func (s *RITControl) SetRITOffset(trx int, offset int) {
	if trx != s.trx {
		return
	}
	s.ValueControl.SetActiveValue(offset)
}

func NewXITControl(key MidiKey, trx int, controlType ControlType, led LED, stepSize int, reverseDirection bool, dynamicMode bool, frequencyRange int, controller XITController) *XITControl {
	set := func(v int) {
		err := controller.SetXITOffset(trx, v)
		if err != nil {
			log.Printf("Cannot change XIT offset: %v", err)
		}
	}
	valueRange := StaticRange{-frequencyRange, frequencyRange}
	return &XITControl{
		ValueControl: NewValueControl(key, controlType, set, valueRange, led, stepSize, reverseDirection, dynamicMode),
		trx:          trx,
	}
}

type XITControl struct {
	ValueControl
	trx int
}

type XITController interface {
	SetXITOffset(trx int, offset int) error
}

func (s *XITControl) SetXITOffset(trx int, offset int) {
	if trx != s.trx {
		return
	}
	s.ValueControl.SetActiveValue(offset)
}
