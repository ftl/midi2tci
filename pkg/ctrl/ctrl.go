package ctrl

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/ftl/tci/client"
)

type MidiKey struct {
	Channel byte
	Key     int8
}

type LED interface {
	Set(key MidiKey, on bool)
}

type Mapping struct {
	Type    MappingType       `json:"type"`
	Channel byte              `json:"channel"`
	Key     int8              `json:"key"`
	TRX     int               `json:"trx"`
	VFO     string            `json:"vfo"`
	Options map[string]string `json:"options"`
}

func (m Mapping) MidiKey() MidiKey {
	return MidiKey{
		Channel: m.Channel,
		Key:     m.Key,
	}
}

func (m Mapping) ValueControlOptions(defaultStepSize int) (controlType ControlType, stepSize int, reverseDirection bool, dynamicMode bool, err error) {
	str := m.Options["control"]
	switch strings.ToLower(str) {
	case "poti":
		controlType = PotiControl
	case "encoder":
		controlType = EncoderControl
	default:
		controlType = PotiControl
	}

	str, ok := m.Options["step"]
	if ok {
		stepSize, err = strconv.Atoi(str)
		if err != nil {
			return
		}
	} else {
		stepSize = defaultStepSize
	}

	if stepSize == 0 {
		stepSize = defaultStepSize
	}

	str = m.Options["direction"]
	reverseDirection = strings.ToLower(str) == "reverse"

	str = m.Options["mode"]
	dynamicMode = strings.ToLower(str) == "dynamic"

	return
}

type MappingType string

type ControlType int

const (
	UnknownControl ControlType = iota
	ButtonControl
	PotiControl
	EncoderControl
)

type ValueRange interface {
	Min() int
	Max() int
	Infinite() bool
}

type StaticRange struct {
	min int
	max int
}

func (r StaticRange) Min() int       { return r.min }
func (r StaticRange) Max() int       { return r.max }
func (r StaticRange) Infinite() bool { return r.min == r.max }

type InfiniteRange struct{}

func (r InfiniteRange) Min() int       { return 0 }
func (r InfiniteRange) Max() int       { return 0 }
func (r InfiniteRange) Infinite() bool { return true }

func RangeTick(r ValueRange) float64 {
	return float64(r.Max()-r.Min()+1) / 128.0
}

func Translate(r ValueRange, value uint8) int {
	if r.Infinite() {
		return int(value)
	}
	return r.Min() + int(float64(value)*RangeTick(r))
}

func Project(r ValueRange, value int) uint8 {
	if r.Infinite() {
		return uint8(value)
	}
	if value < r.Min() {
		return 0
	}
	if value > r.Max() {
		return 0x7f
	}
	p := uint8(math.Ceil(float64(value-r.Min()) / RangeTick(r)))
	return p
}

type ValueControl interface {
	Changed(int)
	SetActiveValue(value int)
	Close()
}

func NewValueControl(controlType ControlType, set func(int), valueRange ValueRange, stepSize int, reverseDirection bool, dynamicMode bool) ValueControl {
	if controlType == EncoderControl {
		return NewEncoder(set, valueRange, stepSize, reverseDirection, dynamicMode)
	} else {
		return NewPoti(set, valueRange)
	}
}

type ControlFactory func(Mapping, LED, *client.Client) (any, ControlType, error)

var Factories = make(map[MappingType]ControlFactory)

func AtoVFO(a string) (client.VFO, error) {
	switch strings.ToUpper(a) {
	case "A", "VFOA":
		return client.VFOA, nil
	case "B", "VFOB":
		return client.VFOB, nil
	default:
		return 0, fmt.Errorf("%s is not a valid VFO, use VFOA or VFOB", a)
	}
}
