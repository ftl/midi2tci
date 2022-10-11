package ctrl

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ftl/tci/client"
)

const (
	FilterMapping      MappingType = "filter"
	FilterWidthMapping MappingType = "filter_width"
)

func init() {
	Factories[FilterMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControlType, error) {
		minFrequency, set, err := m.RequiredIntOption("min")
		if err != nil {
			return nil, ButtonControl, fmt.Errorf("invalid minimum frequency: %w", err)
		}
		if !set {
			return nil, ButtonControl, fmt.Errorf("no minimum frequency configured. Use options[\"min\"]=\"<min frequency in Hz>\" to configure the filter's minimum frequency")
		}

		maxFrequency, set, err := m.RequiredIntOption("max")
		if err != nil {
			return nil, ButtonControl, fmt.Errorf("invalid maximum frequency: %w", err)
		}
		if !set {
			return nil, ButtonControl, fmt.Errorf("no maximum frequency configured. Use options[\"max\"]=\"<max frequency in Hz>\" to configure the filter's maximum frequency")
		}

		mode, ok := m.Options["mode"]
		if ok {
			mode = strings.TrimSpace(strings.ToLower(mode))
		}

		return NewFilterBandButton(m.MidiKey(), m.TRX, minFrequency, maxFrequency, client.Mode(mode), led, tciClient), ButtonControl, nil
	}
	Factories[FilterWidthMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControlType, error) {
		controlType, stepSize, reverseDirection, dynamicMode, err := m.ValueControlOptions(1)
		if err != nil {
			return nil, 0, err
		}

		return NewFilterWidthControl(m.MidiKey(), m.TRX, controlType, led, stepSize, reverseDirection, dynamicMode, tciClient), controlType, nil
	}
}

func NewFilterBandButton(key MidiKey, trx int, bottomFrequency int, topFrequency int, mode client.Mode, led LED, controller RXFilterBandController) *FilterBandButton {
	return &FilterBandButton{
		key:        key,
		trx:        trx,
		led:        led,
		controller: controller,

		bottomFrequency: bottomFrequency,
		topFrequency:    topFrequency,
		mode:            mode,
	}
}

type FilterBandButton struct {
	key        MidiKey
	trx        int
	led        LED
	controller RXFilterBandController

	bottomFrequency int
	topFrequency    int
	mode            client.Mode

	currentBottomFrequency int
	currentTopFrequency    int
	currentMode            client.Mode
}

type RXFilterBandController interface {
	SetRXFilterBand(trx int, min, max int) error
	SetMode(trx int, mode client.Mode) error
}

func (b *FilterBandButton) Pressed() {
	if b.mode != "" {
		err := b.controller.SetMode(b.trx, b.mode)
		if err != nil {
			log.Printf("cannot set mode: %v", err)
		}
	}

	// add a grace period before setting the filter band, otherwise ExpertSDR will restore
	// the last filter band for this mode and overwrite our setting
	time.Sleep(200 * time.Millisecond)

	err := b.controller.SetRXFilterBand(b.trx, b.bottomFrequency, b.topFrequency)
	if err != nil {
		log.Printf("cannot set rx filter band: %v", err)
	}
}

func (b *FilterBandButton) enabled() bool {
	return (b.currentMode == b.mode) &&
		(b.currentBottomFrequency == b.bottomFrequency) &&
		(b.currentTopFrequency == b.topFrequency)
}

func (b *FilterBandButton) SetRXFilterBand(trx int, min, max int) {
	if trx != b.trx {
		return
	}
	b.currentBottomFrequency = min
	b.currentTopFrequency = max
	b.led.SetOn(b.key, b.enabled())
}

func (b *FilterBandButton) SetMode(trx int, mode client.Mode) {
	if trx != b.trx {
		return
	}
	b.currentMode = mode
	b.led.SetOn(b.key, b.enabled())
}

func NewFilterWidthControl(key MidiKey, trx int, controlType ControlType, led LED, stepSize int, reverseDirection bool, dynamicMode bool, controller RXFilterBandController) *FilterWidthControl {
	result := &FilterWidthControl{
		trx: trx,
	}
	set := func(value int) {
		min, max := result.shape.Bounds(value)

		err := controller.SetRXFilterBand(trx, min, max)
		if err != nil {
			log.Printf("Cannot send filter width %d = %d,%d: %v", value, min, max, err)
		}
	}

	result.ValueControl = NewValueControl(key, controlType, set, result, led, stepSize, reverseDirection, dynamicMode)
	return result
}

type FilterWidthControl struct {
	ValueControl
	trx int

	shape   filterShape
	enabled bool
}

func (s *FilterWidthControl) Min() int       { return s.shape.Min() }
func (s *FilterWidthControl) Max() int       { return s.shape.Max() }
func (s *FilterWidthControl) Infinite() bool { return s.shape.Infinite() }

func (s *FilterWidthControl) SetMode(trx int, mode client.Mode) {
	if trx != s.trx {
		return
	}
	s.shape, s.enabled = shapeByMode[mode]
	// TODO s.ValueControl.SetEnabled(s.enabled)
}

func (s *FilterWidthControl) SetRXFilterBand(trx int, min, max int) {
	if trx != s.trx {
		return
	}
	if !s.enabled {
		return
	}

	width := max - min
	if width < 0 {
		width *= -1
	}
	s.ValueControl.SetActiveValue(width)
}

type filterShape struct {
	pivotFrequency int
	minWidth       int
	maxWidth       int
	leftFraction   float64
	rightFraction  float64
}

func (s filterShape) Min() int       { return s.minWidth }
func (s filterShape) Max() int       { return s.maxWidth }
func (s filterShape) Infinite() bool { return false }

func (s filterShape) Width(value int) int {
	const maxControlValue = 127.0
	fraction := float64(value) / maxControlValue
	space := s.maxWidth - s.minWidth
	return s.minWidth + int(float64(space)*fraction)
}

func (s filterShape) Bounds(width int) (int, int) {
	minFrequency := s.pivotFrequency - int(float64(width)*s.leftFraction)
	maxFrequency := s.pivotFrequency + int(float64(width)*s.rightFraction)
	return minFrequency, maxFrequency
}

var shapeByMode = map[client.Mode]filterShape{
	client.ModeCW: {
		minWidth:      50,
		maxWidth:      1000,
		leftFraction:  0.5,
		rightFraction: 0.5,
	},
	client.ModeDIGL: {
		pivotFrequency: -1500,
		minWidth:       100,
		maxWidth:       3000,
		leftFraction:   0.5,
		rightFraction:  0.5,
	},
	client.ModeDIGU: {
		pivotFrequency: 1500,
		minWidth:       100,
		maxWidth:       3000,
		leftFraction:   0.5,
		rightFraction:  0.5,
	},
	client.ModeLSB: {
		pivotFrequency: -70,
		minWidth:       1000,
		maxWidth:       3000,
		leftFraction:   1,
		rightFraction:  0,
	},
	client.ModeUSB: {
		pivotFrequency: 70,
		minWidth:       1000,
		maxWidth:       3000,
		leftFraction:   0,
		rightFraction:  1,
	},
	client.ModeDSB: {
		pivotFrequency: 0,
		minWidth:       1000,
		maxWidth:       6000,
		leftFraction:   0.5,
		rightFraction:  0.5,
	},
	client.ModeAM: {
		pivotFrequency: 0,
		minWidth:       1000,
		maxWidth:       6000,
		leftFraction:   0.5,
		rightFraction:  0.5,
	},
	client.ModeNFM: {
		pivotFrequency: 0,
		minWidth:       1000,
		maxWidth:       25000,
		leftFraction:   0.5,
		rightFraction:  0.5,
	},
	// client.ModeSAM:,
	// client.ModeWFM:,
	// client.ModeSPEC:,
	// client.ModeDRM:,
}
