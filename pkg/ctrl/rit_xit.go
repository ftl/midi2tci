package ctrl

import "log"

const (
	EnableRITMapping MappingType = "enable_rit"
	RITMapping       MappingType = "rit"
	EnableXITMapping MappingType = "enable_xit"
	XITMapping       MappingType = "xit"
)

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

func NewRITSlider(trx int, controller RITController) *RITSlider {
	const tick = float64(1000.0 / 127.0)
	return &RITSlider{
		Slider: NewSlider(
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

type RITSlider struct {
	*Slider
	trx int
}

type RITController interface {
	SetRITOffset(trx int, offset int) error
}

func (s *RITSlider) SetRITOffset(trx int, offset int) {
	if trx != s.trx {
		return
	}
	s.Slider.SetActiveValue(offset)
}

func NewXITSlider(trx int, controller XITController) *XITSlider {
	const tick = float64(1000.0 / 127.0)
	return &XITSlider{
		Slider: NewSlider(
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

type XITSlider struct {
	*Slider
	trx int
}

type XITController interface {
	SetXITOffset(trx int, offset int) error
}

func (s *XITSlider) SetXITOffset(trx int, offset int) {
	if trx != s.trx {
		return
	}
	s.Slider.SetActiveValue(offset)
}
