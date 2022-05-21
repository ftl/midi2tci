package cfg

import (
	"encoding/json"
	"io"
	"os"

	"github.com/ftl/midi2tci/pkg/ctrl"
)

type Configuration struct {
	PortNumber   int            `json:"port_number,omitempty"`
	PortName     string         `json:"port_name,omitempty"`
	TCIAddress   string         `json:"tci_address,omitempty"`
	InitSequence [][]byte       `json:"init_sequence,omitempty"`
	Mappings     []ctrl.Mapping `json:"mappings"`
}

func ReadFile(filename string) (Configuration, error) {
	f, err := os.Open(filename)
	if err != nil {
		return Configuration{}, err
	}
	defer f.Close()

	return Read(f)
}

func Read(r io.Reader) (Configuration, error) {
	var result Configuration

	bytes, err := io.ReadAll(r)
	if err != nil {
		return Configuration{}, err
	}

	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return Configuration{}, err
	}

	return result, nil
}
