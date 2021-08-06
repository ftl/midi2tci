package ctrl

import "log"

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

func NewVolumeSlider(controller VolumeController) *VolumeSlider {
	const tick = float64(60.0 / 127.0)
	return &VolumeSlider{
		Slider: NewSlider(
			func(v int) {
				err := controller.SetVolume(v)
				if err != nil {
					log.Printf("Cannot change volume: %v", err)
				}
			},
			func(v int) int { return -60 + int(float64(v)*tick) },
		),
	}
}

type VolumeSlider struct {
	*Slider
}

type VolumeController interface {
	SetVolume(dB int) error
}

func (s *VolumeSlider) SetVolume(volume int) {
	s.Slider.SetActiveValue(volume)
}
