package ctrl

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ftl/tci/client"
)

type MidiKey struct {
	Channel byte
	Key     byte
}

type LED interface {
	Set(key MidiKey, on bool)
}

type Mapping struct {
	Type    MappingType       `json:"type"`
	Channel byte              `json:"channel"`
	Key     byte              `json:"key"`
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

func (m Mapping) ControlType() ControlType {
	str := m.Options["type"]
	switch strings.ToLower(str) {
	case "button":
		return ButtonControl
	case "poti":
		return PotiControl
	case "encoder":
		return EncoderControl
	default:
		return UnknownControl
	}
}

func (m Mapping) ValueControlOptions(defaultStepSize int) (controlType ControlType, stepSize int, reverseDirection bool, dynamicMode bool, err error) {
	str := m.Options["type"]
	switch strings.ToLower(str) {
	case "button":
		controlType = ButtonControl
	case "poti":
		controlType = PotiControl
	case "encoder":
		controlType = EncoderControl
	default:
		controlType = UnknownControl
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

type ValueControl interface {
	Changed(int)
	SetActiveValue(value int)
	Close()
}

func NewValueControl(controlType ControlType, set func(int), translate func(int) int, stepSize int, reverseDirection bool, dynamicMode bool) ValueControl {
	if controlType == EncoderControl {
		return NewEncoder(set, identity, stepSize, reverseDirection, dynamicMode)
	} else {
		return NewPoti(set, translate)
	}
}

func identity(i int) int { return i }

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
