package ctrl

import (
	"log"

	"github.com/ftl/tci/client"
)

type MidiKey struct {
	Channel byte
	Key     byte
}

type LED interface {
	Set(key MidiKey, on bool)
}

func NewMuteButton(key MidiKey, led LED, muter Muter) *MuteButton {
	return &MuteButton{
		key:   key,
		led:   led,
		muter: muter,
		muted: false,
	}
}

type MuteButton struct {
	key   MidiKey
	led   LED
	muter Muter

	muted bool
}

type Muter interface {
	SetMute(bool) error
}

func (b *MuteButton) Pressed() {
	err := b.muter.SetMute(!b.muted)
	if err != nil {
		log.Print(err)
	}
}

func (b *MuteButton) SetMute(muted bool) {
	b.muted = muted
	b.led.Set(b.key, !muted)
}

func NewVFOWheel(key MidiKey, trx int, vfo client.VFO, controller VFOFrequencyController) *VFOWheel {
	return &VFOWheel{
		key:        key,
		trx:        trx,
		vfo:        vfo,
		controller: controller,
	}
}

type VFOWheel struct {
	key        MidiKey
	trx        int
	vfo        client.VFO
	controller VFOFrequencyController

	frequency      int
	frequencyValid bool
}

type VFOFrequencyController interface {
	SetVFOFrequency(trx int, vfo client.VFO, frequency int) error
}

func (v *VFOWheel) Turned(delta int) {
	v.controller.SetVFOFrequency(v.trx, v.vfo, v.frequency+delta)
}

func (v *VFOWheel) SetVFOFrequency(trx int, vfo client.VFO, frequency int) {
	if trx != v.trx || vfo != v.vfo {
		return
	}
	v.frequency = frequency
	v.frequencyValid = true
}
