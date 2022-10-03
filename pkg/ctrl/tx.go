package ctrl

import (
	"log"

	"github.com/ftl/tci/client"
)

const (
	MOXMapping  MappingType = "mox"
	TuneMapping MappingType = "tune"
)

func init() {
	Factories[MOXMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControlType, error) {
		return NewMOXButton(m.MidiKey(), m.TRX, led, tciClient), ButtonControl, nil
	}
	Factories[TuneMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControlType, error) {
		return NewTuneButton(m.MidiKey(), m.TRX, led, tciClient), ButtonControl, nil
	}
}

func NewMOXButton(key MidiKey, trx int, led LED, enabler MOXEnabler) *MOXButton {
	return &MOXButton{
		key:     key,
		trx:     trx,
		led:     led,
		enabler: enabler,
	}
}

type MOXButton struct {
	key     MidiKey
	trx     int
	led     LED
	enabler MOXEnabler

	enabled bool
}

type MOXEnabler interface {
	SetTX(int, bool, client.SignalSource) error
}

func (b *MOXButton) Pressed() {
	err := b.enabler.SetTX(b.trx, !b.enabled, client.SignalSourceDefault)
	if err != nil {
		log.Print(err)
	}
}

func (b *MOXButton) SetTX(trx int, ptt bool) {
	if trx != b.trx {
		return
	}
	b.enabled = ptt
	b.led.SetFlashing(b.key, ptt)
}

func NewTuneButton(key MidiKey, trx int, led LED, enabler TuneEnabler) *TuneButton {
	return &TuneButton{
		key:     key,
		trx:     trx,
		led:     led,
		enabler: enabler,
	}
}

type TuneButton struct {
	key     MidiKey
	trx     int
	led     LED
	enabler TuneEnabler

	enabled bool
}

type TuneEnabler interface {
	SetTune(int, bool) error
}

func (b *TuneButton) Pressed() {
	err := b.enabler.SetTune(b.trx, !b.enabled)
	if err != nil {
		log.Print(err)
	}
}

func (b *TuneButton) SetTune(trx int, ptt bool) {
	if trx != b.trx {
		return
	}
	b.enabled = ptt
	b.led.SetOn(b.key, ptt)
}
