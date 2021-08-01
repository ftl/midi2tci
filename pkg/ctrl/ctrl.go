package ctrl

import (
	"log"
	"time"

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

func NewRXChannelEnableButton(key MidiKey, trx int, vfo client.VFO, led LED, rxChannelEnabler RXChannelEnabler) *RXChannelEnableButton {
	return &RXChannelEnableButton{
		key:              key,
		trx:              trx,
		vfo:              vfo,
		led:              led,
		rxChannelEnabler: rxChannelEnabler,
	}
}

type RXChannelEnableButton struct {
	key              MidiKey
	trx              int
	vfo              client.VFO
	led              LED
	rxChannelEnabler RXChannelEnabler

	enabled bool
}

type RXChannelEnabler interface {
	SetRXChannelEnable(int, client.VFO, bool) error
}

func (b *RXChannelEnableButton) Pressed() {
	err := b.rxChannelEnabler.SetRXChannelEnable(b.trx, b.vfo, !b.enabled)
	if err != nil {
		log.Print(err)
	}
}

func (b *RXChannelEnableButton) SetRXChannelEnable(trx int, vfo client.VFO, enabled bool) {
	if trx != b.trx || vfo != b.vfo {
		return
	}
	b.enabled = enabled
	b.led.Set(b.key, enabled)
}

func NewSplitEnableButton(key MidiKey, trx int, led LED, splitEnabler SplitEnabler) *SplitEnableButton {
	return &SplitEnableButton{
		key:          key,
		trx:          trx,
		led:          led,
		splitEnabler: splitEnabler,
	}
}

type SplitEnableButton struct {
	key          MidiKey
	trx          int
	led          LED
	splitEnabler SplitEnabler

	enabled bool
}

type SplitEnabler interface {
	SetSplitEnable(int, bool) error
}

func (b *SplitEnableButton) Pressed() {
	err := b.splitEnabler.SetSplitEnable(b.trx, !b.enabled)
	if err != nil {
		log.Print(err)
	}
}

func (b *SplitEnableButton) SetSplitEnable(trx int, enabled bool) {
	if trx != b.trx {
		return
	}
	b.enabled = enabled
	b.led.Set(b.key, enabled)
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

func NewSyncVFOFrequencyButton(srcTrx int, srcVFO client.VFO, dstTrx int, dstVFO client.VFO, controller VFOFrequencyController, provider VFOFrequencyProvider) *SyncVFOFrequencyButton {
	return &SyncVFOFrequencyButton{
		srcTrx:     srcTrx,
		srcVFO:     srcVFO,
		dstTrx:     dstTrx,
		dstVFO:     dstVFO,
		controller: controller,
		provider:   provider,
	}
}

type SyncVFOFrequencyButton struct {
	srcTrx     int
	srcVFO     client.VFO
	dstTrx     int
	dstVFO     client.VFO
	controller VFOFrequencyController
	provider   VFOFrequencyProvider
}

type VFOFrequencyProvider interface {
	VFOFrequency(trx int, vfo client.VFO) (int, error)
}

func (b *SyncVFOFrequencyButton) Pressed() {
	frequency, err := b.provider.VFOFrequency(b.srcTrx, b.srcVFO)
	if err != nil {
		log.Printf("Cannot read VFO frequency: %v", err)
		return
	}
	err = b.controller.SetVFOFrequency(b.dstTrx, b.dstVFO, frequency)
}

func NewVFOWheel(key MidiKey, trx int, vfo client.VFO, controller VFOFrequencyController) *VFOWheel {
	result := &VFOWheel{
		key:        key,
		trx:        trx,
		vfo:        vfo,
		controller: controller,
		frequency:  make(chan int, 1000),
		turns:      make(chan int, 1000),
		closed:     make(chan struct{}),
	}

	go func() {
		defer close(result.closed)
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		accumulatedTurns := 0
		turning := false
		frequency := 0
		for {
			select {
			case turns, valid := <-result.turns:
				if !valid {
					return
				}
				accumulatedTurns += turns
				turning = frequency > 0
			case f := <-result.frequency:
				if !turning {
					frequency = f
				}
			case <-ticker.C:
				if accumulatedTurns == 0 {
					turning = false
				} else if accumulatedTurns != 0 && frequency != 0 {
					frequency = frequency + int(float64(accumulatedTurns)*1.8)
					err := result.controller.SetVFOFrequency(result.trx, result.vfo, frequency)
					if err != nil {
						log.Printf("Cannot change frequency to %d: %v", result.frequency, err)
					}
					accumulatedTurns = 0
				}
			}
		}
	}()

	return result
}

type VFOWheel struct {
	key        MidiKey
	trx        int
	vfo        client.VFO
	controller VFOFrequencyController

	frequency chan int
	turns     chan int
	closed    chan struct{}
}

type VFOFrequencyController interface {
	SetVFOFrequency(trx int, vfo client.VFO, frequency int) error
}

func (w *VFOWheel) Close() {
	select {
	case <-w.closed:
		return
	default:
		close(w.turns)
		<-w.closed
	}
}

func (w *VFOWheel) Turned(turns int) {
	w.turns <- turns
}

func (w *VFOWheel) SetVFOFrequency(trx int, vfo client.VFO, frequency int) {
	if trx != w.trx || vfo != w.vfo {
		return
	}
	w.frequency <- frequency
}

func NewSlider(set func(int), translate func(int) int) *Slider {
	result := &Slider{
		set:           set,
		translate:     translate,
		selectedValue: make(chan int, 1000),
		activeValue:   make(chan int, 1000),
		closed:        make(chan struct{}),
	}

	result.start()

	return result
}

type Slider struct {
	set           func(int)
	translate     func(int) int
	activeValue   chan int
	selectedValue chan int
	closed        chan struct{}
}

func (s *Slider) start() {
	tx := make(chan int)
	go func() {
		for {
			select {
			case <-s.closed:
				return
			case value := <-tx:
				s.set(value)
			}
		}
	}()

	go func() {
		defer close(s.closed)
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		activeValue := 0
		selectedValue := 0
		pending := false

		for {
			select {
			case value, valid := <-s.activeValue:
				if !valid {
					return
				}
				activeValue = value
				if !pending {
					selectedValue = activeValue
				}
			case value, valid := <-s.selectedValue:
				if !valid {
					return
				}
				selectedValue = value

				if activeValue == selectedValue {
					continue
				}

				select {
				case tx <- selectedValue:
					pending = false
				default:
					pending = true
				}
			case <-ticker.C:
				if activeValue == selectedValue {
					pending = false
					continue
				}

				select {
				case tx <- selectedValue:
					pending = false
				default:
					pending = true
				}
			}
		}
	}()
}

func (s *Slider) Close() {
	select {
	case <-s.closed:
		return
	default:
		close(s.activeValue)
		close(s.selectedValue)
		<-s.closed
	}
}

func (s *Slider) Changed(value int) {
	s.selectedValue <- s.translate(value)
}

func (s *Slider) SetActiveValue(value int) {
	s.activeValue <- value
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

func NewRXVolumeSlider(trx int, vfo client.VFO, controller RXVolumeController) *RXVolumeSlider {
	const tick = float64(60.0 / 127.0)
	return &RXVolumeSlider{
		Slider: NewSlider(
			func(v int) {
				err := controller.SetRXVolume(trx, vfo, v)
				if err != nil {
					log.Printf("Cannot change RX volume: %v", err)
				}
			},
			func(v int) int { return -60 + int(float64(v)*tick) },
		),
		trx: trx,
		vfo: vfo,
	}
}

type RXVolumeSlider struct {
	*Slider
	trx int
	vfo client.VFO
}

type RXVolumeController interface {
	SetRXVolume(trx int, vfo client.VFO, dB int) error
}

func (s *RXVolumeSlider) SetRXVolume(trx int, vfo client.VFO, volume int) {
	if trx != s.trx || vfo != s.vfo {
		return
	}
	s.Slider.SetActiveValue(volume)
}

func NewRXBalanceSlider(trx int, vfo client.VFO, controller RXBalanceController) *RXBalanceSlider {
	const tick = float64(80.0 / 127.0)
	return &RXBalanceSlider{
		Slider: NewSlider(
			func(v int) {
				err := controller.SetRXBalance(trx, vfo, v)
				if err != nil {
					log.Printf("Cannot change RX balance: %v", err)
				}
			},
			func(v int) int { return -40 + int(float64(v)*tick) },
		),
		trx: trx,
		vfo: vfo,
	}
}

type RXBalanceSlider struct {
	*Slider
	trx int
	vfo client.VFO
}

type RXBalanceController interface {
	SetRXBalance(trx int, vfo client.VFO, dB int) error
}

func (s *RXBalanceSlider) SetRXBalance(trx int, vfo client.VFO, balance int) {
	if trx != s.trx || vfo != s.vfo {
		return
	}
	s.Slider.SetActiveValue(balance)
}

func NewRXMixer(trx int, controller RXMixController) *RXMixer {
	const volumeTick = float64(60.0 / 127.0)
	const balanceTick = float64(80.0 / 127.0)
	return &RXMixer{
		vfoAVolume: NewSlider(
			func(v int) {
				err := controller.SetRXVolume(trx, client.VFOA, v)
				if err != nil {
					log.Printf("Cannot change RX volume: %v", err)
				}
			},
			func(v int) int { return -60 + int(float64(v)*volumeTick) },
		),
		vfoABalance: NewSlider(
			func(v int) {
				err := controller.SetRXBalance(trx, client.VFOA, v)
				if err != nil {
					log.Printf("Cannot change RX balance: %v", err)
				}
			},
			func(v int) int { return -40 + int(float64(v)*balanceTick) },
		),
		vfoBVolume: NewSlider(
			func(v int) {
				err := controller.SetRXVolume(trx, client.VFOB, v)
				if err != nil {
					log.Printf("Cannot change RX volume: %v", err)
				}
			},
			func(v int) int { return -60 + int(float64(v)*volumeTick) },
		),
		vfoBBalance: NewSlider(
			func(v int) {
				err := controller.SetRXBalance(trx, client.VFOB, v)
				if err != nil {
					log.Printf("Cannot change RX balance: %v", err)
				}
			},
			func(v int) int { return -40 + int(float64(v)*balanceTick) },
		),
		trx: trx,
	}
}

type RXMixer struct {
	vfoAVolume  *Slider
	vfoABalance *Slider
	vfoBVolume  *Slider
	vfoBBalance *Slider
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
