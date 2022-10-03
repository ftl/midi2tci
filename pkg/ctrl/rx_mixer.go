package ctrl

import (
	"log"

	"github.com/ftl/tci/client"
)

const MixerMapping MappingType = "rx_mixer"

func init() {
	Factories[MixerMapping] = func(m Mapping, _ LED, tciClient *client.Client) (interface{}, ControlType, error) {
		return NewRXMixer(m.TRX, tciClient), PotiControl, nil
	}
}

func NewRXMixer(trx int, controller RXMixController) *RXMixer {
	volumeRange := StaticRange{-60, 0}
	balanceRange := StaticRange{-40, 40}
	return &RXMixer{
		vfoAVolume: NewPoti(
			MidiKey{},
			func(v int) {
				err := controller.SetRXVolume(trx, client.VFOA, v)
				if err != nil {
					log.Printf("Cannot change RX volume: %v", err)
				}
			},
			volumeRange,
			nil,
		),
		vfoABalance: NewPoti(
			MidiKey{},
			func(v int) {
				err := controller.SetRXBalance(trx, client.VFOA, v)
				if err != nil {
					log.Printf("Cannot change RX balance: %v", err)
				}
			},
			balanceRange,
			nil,
		),
		vfoBVolume: NewPoti(
			MidiKey{},
			func(v int) {
				err := controller.SetRXVolume(trx, client.VFOB, v)
				if err != nil {
					log.Printf("Cannot change RX volume: %v", err)
				}
			},
			volumeRange,
			nil,
		),
		vfoBBalance: NewPoti(
			MidiKey{},
			func(v int) {
				err := controller.SetRXBalance(trx, client.VFOB, v)
				if err != nil {
					log.Printf("Cannot change RX balance: %v", err)
				}
			},
			balanceRange,
			nil,
		),
		trx: trx,
	}
}

type RXMixer struct {
	vfoAVolume  ValueControl
	vfoABalance ValueControl
	vfoBVolume  ValueControl
	vfoBBalance ValueControl
	trx         int
}

type RXMixController interface {
	SetRXVolume(trx int, vfo client.VFO, dB int) error
	SetRXBalance(trx int, vfo client.VFO, dB int) error
}

func (s *RXMixer) Close() {
	s.vfoAVolume.Close()
	s.vfoABalance.Close()
	s.vfoBVolume.Close()
	s.vfoBBalance.Close()
}

func (s *RXMixer) Changed(value int) {
	const (
		min    = 0x00
		max    = 0x7f
		right  = 0x7f
		center = 0x40
		left   = 0x00
	)
	var (
		vfoAVolume  int
		vfoABalance int
		vfoBVolume  int
		vfoBBalance int
	)
	if value == center {
		vfoAVolume = max
		vfoBVolume = max
		vfoABalance = left
		vfoBBalance = right
	} else if value < center {
		vfoAVolume = max
		vfoBVolume = max - (center-value)*2
		vfoABalance = center - value
		vfoBBalance = right
	} else {
		vfoAVolume = max - (value-center)*2
		vfoBVolume = max
		vfoABalance = left
		vfoBBalance = right - (value - center)
	}

	s.vfoAVolume.Changed(vfoAVolume)
	s.vfoABalance.Changed(vfoABalance)
	s.vfoBVolume.Changed(vfoBVolume)
	s.vfoBBalance.Changed(vfoBBalance)
}

func (s *RXMixer) SetRXVolume(trx int, vfo client.VFO, volume int) {
	if trx != s.trx {
		return
	}
	switch vfo {
	case client.VFOA:
		s.vfoAVolume.SetActiveValue(volume)
	case client.VFOB:
		s.vfoBVolume.SetActiveValue(volume)
	}
}

func (s *RXMixer) SetRXBalance(trx int, vfo client.VFO, balance int) {
	if trx != s.trx {
		return
	}
	switch vfo {
	case client.VFOA:
		s.vfoABalance.SetActiveValue(balance)
	case client.VFOB:
		s.vfoBBalance.SetActiveValue(balance)
	}
}
