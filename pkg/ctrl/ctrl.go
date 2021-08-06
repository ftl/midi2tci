package ctrl

import (
	"fmt"
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

type MappingType string

type ControllerType int

const (
	ButtonController ControllerType = iota
	SliderController
	WheelController
)

type ControllerFactory func(Mapping, LED, *client.Client) (interface{}, ControllerType, error)

var Factories = make(map[MappingType]ControllerFactory)

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
