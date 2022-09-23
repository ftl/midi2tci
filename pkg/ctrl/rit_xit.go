package ctrl

import (
	"log"

	"github.com/ftl/tci/client"
)

const (
	EnableRITMapping MappingType = "enable_rit"
	RITMapping       MappingType = "rit"
	EnableXITMapping MappingType = "enable_xit"
	XITMapping       MappingType = "xit"
)

func init() {
	Factories[EnableRITMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControlType, error) {
		return NewRITEnableButton(m.MidiKey(), m.TRX, led, tciClient), ButtonControl, nil
	}
	Factories[RITMapping] = func(m Mapping, _ LED, tciClient *client.Client) (interface{}, ControlType, error) {
		return NewRITControl(m.TRX, tciClient), PotiControl, nil
	}
	Factories[EnableXITMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControlType, error) {
		return NewXITEnableButton(m.MidiKey(), m.TRX, led, tciClient), ButtonControl, nil
	}
	Factories[XITMapping] = func(m Mapping, _ LED, tciClient *client.Client) (interface{}, ControlType, error) {
		return NewXITControl(m.TRX, tciClient), PotiControl, nil
	}
}

func NewRITEnableButton(key MidiKey, trx int, led LED, ritEnabler RITEnabler) *RITEnableButton {
	return &RITEnableButton{
		key:        key,
		trx:        trx,
		led:        led,
		ritEnabler: ritEnabler,
	}
}

type RITEnableButton struct {
	key        MidiKey
	trx        int
	led        LED
	ritEnabler RITEnabler

	enabled bool
}

type RITEnabler interface {
	SetRITEnable(int, bool) error
}

func (b *RITEnableButton) Pressed() {
	err := b.ritEnabler.SetRITEnable(b.trx, !b.enabled)
	if err != nil {
		log.Print(err)
	}
}

func (b *RITEnableButton) SetRITEnable(trx int, enabled bool) {
	if trx != b.trx {
		return
	}
	b.enabled = enabled
	b.led.Set(b.key, enabled)
}

func NewXITEnableButton(key MidiKey, trx int, led LED, xitEnabler XITEnabler) *XITEnableButton {
	return &XITEnableButton{
		key:        key,
		trx:        trx,
		led:        led,
		xitEnabler: xitEnabler,
	}
}

type XITEnableButton struct {
	key        MidiKey
	trx        int
	led        LED
	xitEnabler XITEnabler

	enabled bool
}

type XITEnabler interface {
	SetXITEnable(int, bool) error
}

func (b *XITEnableButton) Pressed() {
	err := b.xitEnabler.SetXITEnable(b.trx, !b.enabled)
	if err != nil {
		log.Print(err)
	}
}

func (b *XITEnableButton) SetXITEnable(trx int, enabled bool) {
	if trx != b.trx {
		return
	}
	b.enabled = enabled
	b.led.Set(b.key, enabled)
}

func NewRITControl(trx int, controller RITController) *RITControl {
	const tick = float64(1000.0 / 127.0)
	return &RITControl{
		ValueControl: NewPoti(
			func(v int) {
				err := controller.SetRITOffset(trx, v)
				if err != nil {
					log.Printf("Cannot change RIT offset: %v", err)
				}
			},
			func(v int) int {
				if v == 0x40 {
					return 0
				}
				return -500 + int(float64(v)*tick)
			},
		),
		trx: trx,
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

func NewXITControl(trx int, controller XITController) *XITControl {
	const tick = float64(1000.0 / 127.0)
	return &XITControl{
		ValueControl: NewPoti(
			func(v int) {
				err := controller.SetXITOffset(trx, v)
				if err != nil {
					log.Printf("Cannot change XIT offset: %v", err)
				}
			},
			func(v int) int {
				if v == 0x40 {
					return 0
				}
				return -500 + int(float64(v)*tick)
			},
		),
		trx: trx,
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
