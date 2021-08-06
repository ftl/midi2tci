package cfg

import (
	"github.com/ftl/midi2tci/pkg/ctrl"
)

type Configuration struct {
	PortNumber int            `json:"port_number,omitempty"`
	PortName   string         `json:"port_name,omitempty"`
	Mappings   []ctrl.Mapping `json:"mappings"`
}
