package ctrl

import (
	"fmt"
	"log"
	"strconv"

	"github.com/ftl/tci/client"
)

const (
	EnableSplitMapping      MappingType = "enable_split"
	SyncVFOFrequencyMapping MappingType = "sync_vfo_frequency"
)

func init() {
	Factories[EnableSplitMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControllerType, error) {
		return NewSplitEnableButton(m.MidiKey(), m.TRX, led, tciClient), ButtonController, nil
	}
	Factories[SyncVFOFrequencyMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControllerType, error) {
		vfo, err := AtoVFO(m.VFO)
		if err != nil {
			return nil, 0, err
		}

		srcTRXStr, ok := m.Options["src_trx"]
		if !ok {
			return nil, ButtonController, fmt.Errorf("no source TRX configured. Use options[\"src_trx\"]=\"<source TRX>\" to configure the source TRX")
		}
		srcTRX, err := strconv.Atoi(srcTRXStr)
		if err != nil {
			return nil, ButtonController, fmt.Errorf("invalid source TRX %s: %v", srcTRXStr, err)
		}

		srcVFOStr, ok := m.Options["src_vfo"]
		if !ok {
			return nil, ButtonController, fmt.Errorf("no source VFO configured. Use options[\"src_vfo\"]=\"<source VFO>\" to configure the source VFO")
		}
		srcVFO, err := AtoVFO(srcVFOStr)
		if err != nil {
			return nil, ButtonController, fmt.Errorf("invalid source VFO %s: %v", srcVFOStr, err)
		}

		offset := 0
		offsetStr, ok := m.Options["offset"]
		if ok {
			offset, err = strconv.Atoi(offsetStr)
			if err != nil {
				return nil, ButtonController, fmt.Errorf("invalid offset %s: %v", offsetStr, err)
			}
		}

		return NewSyncVFOFrequencyButton(srcTRX, srcVFO, m.TRX, vfo, offset, tciClient, tciClient), ButtonController, nil
	}
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

func NewSyncVFOFrequencyButton(srcTrx int, srcVFO client.VFO, dstTrx int, dstVFO client.VFO, offset int, controller VFOFrequencyController, provider VFOFrequencyProvider) *SyncVFOFrequencyButton {
	return &SyncVFOFrequencyButton{
		srcTrx:     srcTrx,
		srcVFO:     srcVFO,
		dstTrx:     dstTrx,
		dstVFO:     dstVFO,
		offset:     offset,
		controller: controller,
		provider:   provider,
	}
}

type SyncVFOFrequencyButton struct {
	srcTrx int
	srcVFO client.VFO
	dstTrx int
	dstVFO client.VFO

	offset int

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
	err = b.controller.SetVFOFrequency(b.dstTrx, b.dstVFO, frequency+b.offset)
	if err != nil {
		log.Printf("Cannot write VFO frequency: %v", err)
	}
}
