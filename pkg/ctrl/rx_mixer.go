package ctrl

import (
	"log"

	"github.com/ftl/tci/client"
)

const (
	MixerMapping    MappingType = "rx_mixer"
	SetMixerMapping MappingType = "set_rx_mixer"
)

func init() {
	Factories[MixerMapping] = func(m Mapping, _ LED, tciClient *client.Client) (interface{}, ControlType, error) {
		return NewRXMixer(m.TRX, tciClient), PotiControl, nil
	}
	Factories["experimental_"+MixerMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControlType, error) {
		controlType, stepSize, reverseDirection, dynamicMode, err := m.ValueControlOptions(1)
		if err != nil {
			return nil, 0, err
		}
		return NewRXMixer2(m.MidiKey(), m.TRX, controlType, led, stepSize, reverseDirection, dynamicMode, tciClient), controlType, nil
	}
	Factories[SetMixerMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControlType, error) {
		volumeA, err := m.IntOption("volume_a", 0)
		if err != nil {
			return nil, 0, err
		}
		volumeB, err := m.IntOption("volume_b", 0)
		if err != nil {
			return nil, 0, err
		}
		balanceA, err := m.IntOption("balance_a", 0)
		if err != nil {
			return nil, 0, err
		}
		balanceB, err := m.IntOption("balance_b", 0)
		if err != nil {
			return nil, 0, err
		}

		return NewSetRXMixerButton(m.MidiKey(), m.TRX, led, volumeA, volumeB, balanceA, balanceB, tciClient), ButtonControl, nil
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

var (
	volumeRange  = StaticRange{-60, 0}
	balanceRange = StaticRange{-40, 40}
)

func NewRXMixer2(key MidiKey, trx int, controlType ControlType, led LED, stepSize int, reverseDirection bool, dynamicMode bool, controller RXMixController) *RXMixer2 {
	result := &RXMixer2{
		trx:        trx,
		controller: controller,
	}
	valueRange := StaticRange{0x00, 0x7f}
	result.ValueControl = NewValueControl(key, controlType, result.set, valueRange, led, stepSize, reverseDirection, dynamicMode)
	return result
}

// experimental
type RXMixer2 struct {
	ValueControl
	trx        int
	controller RXMixController

	targetValue int

	volumeA  int
	volumeB  int
	balanceA int
	balanceB int

	currentVolumeA  int
	currentVolumeB  int
	currentBalanceA int
	currentBalanceB int
}

func (m *RXMixer2) set(value int) {
	m.targetValue = value

	const (
		min    = 0x00
		max    = 0x7f
		right  = 0x7f
		center = 0x40
		left   = 0x00
	)
	if value == center {
		m.volumeA = Translate(volumeRange, max)
		m.volumeB = Translate(volumeRange, max)
		m.balanceA = Translate(balanceRange, left)
		m.balanceB = Translate(balanceRange, right)
	} else if value < center {
		m.volumeA = Translate(volumeRange, max)
		m.volumeB = Translate(volumeRange, uint8(max-(center-value)*2))
		m.balanceA = Translate(balanceRange, uint8(center-value))
		m.balanceB = Translate(balanceRange, right)
	} else {
		m.volumeA = Translate(volumeRange, uint8(max-(value-center)*2))
		m.volumeB = Translate(volumeRange, max)
		m.balanceA = Translate(balanceRange, left)
		m.balanceB = Translate(balanceRange, uint8(right-(value-center)))
	}

	err := m.controller.SetRXVolume(m.trx, client.VFOA, m.volumeA)
	if err != nil {
		log.Print(err)
	}
	err = m.controller.SetRXBalance(m.trx, client.VFOA, m.balanceA)
	if err != nil {
		log.Print(err)
	}
	err = m.controller.SetRXVolume(m.trx, client.VFOB, m.volumeB)
	if err != nil {
		log.Print(err)
	}
	err = m.controller.SetRXBalance(m.trx, client.VFOB, m.balanceB)
	if err != nil {
		log.Print(err)
	}

	log.Printf("VALUE: %d VolA: %d VolB: %d BalA: %d BalB: %d", value, m.volumeA, m.volumeB, m.balanceA, m.balanceB)
}

func (m *RXMixer2) targetReached() bool {
	return (m.volumeA == m.currentVolumeA) && (m.volumeB == m.currentVolumeB) && (m.balanceA == m.currentBalanceA) && (m.balanceB == m.currentBalanceB)
}

func (m *RXMixer2) SetRXVolume(trx int, vfo client.VFO, volume int) {
	if trx != m.trx {
		return
	}
	switch vfo {
	case client.VFOA:
		m.currentVolumeA = volume
	case client.VFOB:
		m.currentVolumeB = volume
	}
	if m.targetReached() {
		log.Printf("target reached: %d", m.targetValue)
		m.SetActiveValue(m.targetValue)
	}
}

func (m *RXMixer2) SetRXBalance(trx int, vfo client.VFO, balance int) {
	if trx != m.trx {
		return
	}
	switch vfo {
	case client.VFOA:
		m.currentBalanceA = balance
	case client.VFOB:
		m.currentBalanceB = balance
	}
	if m.targetReached() {
		log.Printf("target reached: %d", m.targetValue)
		m.SetActiveValue(m.targetValue)
	}
}

func NewSetRXMixerButton(key MidiKey, trx int, led LED, volumeA, volumeB, balanceA, balanceB int, controller RXMixController) *SetRXMixerButton {
	return &SetRXMixerButton{
		key:        key,
		trx:        trx,
		led:        led,
		controller: controller,
		volumeA:    volumeA,
		volumeB:    volumeB,
		balanceA:   balanceA,
		balanceB:   balanceB,
	}
}

type SetRXMixerButton struct {
	key        MidiKey
	trx        int
	led        LED
	controller RXMixController

	volumeA  int
	volumeB  int
	balanceA int
	balanceB int

	currentVolumeA  int
	currentVolumeB  int
	currentBalanceA int
	currentBalanceB int
}

func (b *SetRXMixerButton) Pressed() {
	err := b.controller.SetRXVolume(b.trx, client.VFOA, b.volumeA)
	if err != nil {
		log.Print(err)
	}
	err = b.controller.SetRXBalance(b.trx, client.VFOA, b.balanceA)
	if err != nil {
		log.Print(err)
	}
	err = b.controller.SetRXVolume(b.trx, client.VFOB, b.volumeB)
	if err != nil {
		log.Print(err)
	}
	err = b.controller.SetRXBalance(b.trx, client.VFOB, b.balanceB)
	if err != nil {
		log.Print(err)
	}
}

func (b *SetRXMixerButton) enabled() bool {
	return (b.volumeA == b.currentVolumeA) && (b.volumeB == b.currentVolumeB) && (b.balanceA == b.currentBalanceA) && (b.balanceB == b.currentBalanceB)
}

func (b *SetRXMixerButton) SetRXVolume(trx int, vfo client.VFO, volume int) {
	if trx != b.trx {
		return
	}
	switch vfo {
	case client.VFOA:
		b.currentVolumeA = volume
	case client.VFOB:
		b.currentVolumeB = volume
	}
	b.led.SetOn(b.key, b.enabled())
}

func (b *SetRXMixerButton) SetRXBalance(trx int, vfo client.VFO, balance int) {
	if trx != b.trx {
		return
	}
	switch vfo {
	case client.VFOA:
		b.currentBalanceA = balance
	case client.VFOB:
		b.currentBalanceB = balance
	}
	b.led.SetOn(b.key, b.enabled())
}
