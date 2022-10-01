package ctrl

import (
	"log"

	"github.com/ftl/tci/client"
)

const (
	MuteMapping   MappingType = "mute"
	VolumeMapping MappingType = "volume"
)

func init() {
	Factories[MuteMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControlType, error) {
		return NewMuteButton(m.MidiKey(), led, tciClient), ButtonControl, nil
	}
	Factories[VolumeMapping] = func(m Mapping, _ LED, tciClient *client.Client) (interface{}, ControlType, error) {
		controlType, stepSize, reverseDirection, dynamicMode, err := m.ValueControlOptions(1)
		if err != nil {
			return nil, 0, err
		}
		return NewVolumeControl(controlType, stepSize, reverseDirection, dynamicMode, tciClient), controlType, nil
	}
}

func NewMuteButton(key MidiKey, led LED, muter Muter) *MuteButton {
	return &MuteButton{
		key:   key,
		led:   led,
		muter: muter,
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

func NewVolumeControl(controlType ControlType, stepSize int, reverseDirection bool, dynamicMode bool, controller VolumeController) *VolumeControl {
	set := func(v int) {
		err := controller.SetVolume(v)
		if err != nil {
			log.Printf("Cannot change volume: %v", err)
		}
	}
	valueRange := StaticRange{-60, 0}

	return &VolumeControl{
		ValueControl: NewValueControl(controlType, set, valueRange, stepSize, reverseDirection, dynamicMode),
	}
}

type VolumeControl struct {
	ValueControl
}

type VolumeController interface {
	SetVolume(dB int) error
}

func (s *VolumeControl) SetVolume(volume int) {
	s.ValueControl.SetActiveValue(volume)
}
