package ctrl

import (
	"fmt"
	"log"
	"strconv"

	"github.com/ftl/tci/client"
)

const (
	FilterMapping      MappingType = "filter"
	FilterWidthMapping MappingType = "filter_width"
)

func init() {
	Factories[FilterMapping] = func(m Mapping, led LED, tciClient *client.Client) (interface{}, ControllerType, error) {
		minFrequencyStr, ok := m.Options["min"]
		if !ok {
			return nil, ButtonController, fmt.Errorf("no minimum frequency configured. Use options[\"min\"]=\"<min frequency in Hz>\" to configure the filter's minimum frequency")
		}
		minFrequency, err := strconv.Atoi(minFrequencyStr)
		if err != nil {
			return nil, ButtonController, fmt.Errorf("invalid minimum frequency %s: %v", minFrequencyStr, err)
		}

		maxFrequencyStr, ok := m.Options["max"]
		if !ok {
			return nil, ButtonController, fmt.Errorf("no maximum frequency configured. Use options[\"max\"]=\"<max frequency in Hz>\" to configure the filter's maximum frequency")
		}
		maxFrequency, err := strconv.Atoi(maxFrequencyStr)
		if err != nil {
			return nil, ButtonController, fmt.Errorf("invalid maximum frequency %s: %v", maxFrequencyStr, err)
		}

		return NewFilterBandButton(m.MidiKey(), m.TRX, minFrequency, maxFrequency, led, tciClient), ButtonController, nil
	}
	Factories[FilterWidthMapping] = func(m Mapping, _ LED, tciClient *client.Client) (interface{}, ControllerType, error) {
		return NewFilterWidthSlider(m.TRX, tciClient), SliderController, nil
	}
}

func NewFilterBandButton(key MidiKey, trx int, bottomFrequency int, topFrequency int, led LED, controller RXFilterBandController) *FilterBandButton {
	return &FilterBandButton{
		key:        key,
		trx:        trx,
		led:        led,
		controller: controller,

		bottomFrequency: bottomFrequency,
		topFrequency:    topFrequency,
	}
}

type FilterBandButton struct {
	key        MidiKey
	trx        int
	led        LED
	controller RXFilterBandController

	bottomFrequency int
	topFrequency    int

	enabled bool
}

type RXFilterBandController interface {
	SetRXFilterBand(trx int, min, max int) error
}

func (b *FilterBandButton) Pressed() {
	err := b.controller.SetRXFilterBand(b.trx, b.bottomFrequency, b.topFrequency)
	if err != nil {
		log.Print(err)
	}
}

func (b *FilterBandButton) SetRXFilterBand(trx int, min, max int) {
	if trx != b.trx {
		return
	}
	b.enabled = (min == b.bottomFrequency) && (max == b.topFrequency)
	b.led.Set(b.key, b.enabled)
}

func NewFilterWidthSlider(trx int, controller RXFilterBandController) *FilterWidthSlider {
	result := &FilterWidthSlider{
		trx: trx,
	}
	result.Slider = NewSlider(func(value int) {
		// set
		min, max := result.shape.Bounds(value)

		err := controller.SetRXFilterBand(trx, min, max)
		if err != nil {
			log.Printf("Cannot send filter width %d = %d,%d: %v", value, min, max, err)
		}
	}, func(value int) int {
		// translate
		return result.shape.Width(value)
	})
	return result
}

type FilterWidthSlider struct {
	*Slider
	trx int

	shape   filterShape
	enabled bool
}

func (s *FilterWidthSlider) SetMode(trx int, mode client.Mode) {
	if trx != s.trx {
		return
	}
	s.shape, s.enabled = shapeByMode[mode]
	// TODO s.Slider.SetEnabled(s.enabled)
}

func (s *FilterWidthSlider) SetRXFilterBand(trx int, min, max int) {
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
	s.Slider.SetActiveValue(width)
}

type filterShape struct {
	pivotFrequency int
	minWidth       int
	maxWidth       int
	leftFraction   float64
	rightFraction  float64
}

func (s filterShape) Width(value int) int {
	const maxSliderValue = 127.0
	fraction := float64(value) / maxSliderValue
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
