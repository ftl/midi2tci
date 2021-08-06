package ctrl

import (
	"log"

	"github.com/ftl/tci/client"
)

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
