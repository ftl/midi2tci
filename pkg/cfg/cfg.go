package cfg

import (
	"github.com/ftl/midi2tci/pkg/ctrl"
)

type Configuration struct {
	PortNumber int            `json:"port_number,omitempty"`
	PortName   string         `json:"port_name,omitempty"`
	Mappings   []ctrl.Mapping `json:"mappings"`
}

// var factories = map[MappingType]ControllerFactory{
// 	VFOMapping: func(m Mapping, _ ctrl.LED, tciClient *client.Client) (interface{}, ControllerType, error) {
// 		vfo, err := AtoVFO(m.VFO)
// 		if err != nil {
// 			return nil, 0, err
// 		}
// 		return ctrl.NewVFOWheel(m.MidiKey(), m.TRX, vfo, tciClient), WheelController, nil
// 	},
// 	MixerMapping: func(m Mapping, _ ctrl.LED, tciClient *client.Client) (interface{}, ControllerType, error) {
// 		return ctrl.NewRXMixer(m.TRX, tciClient), SliderController, nil
// 	},
// 	MuteMapping: func(m Mapping, led ctrl.LED, tciClient *client.Client) (interface{}, ControllerType, error) {
// 		return ctrl.NewMuteButton(m.MidiKey(), led, tciClient), ButtonController, nil
// 	},
// 	VolumeMapping: func(_ Mapping, _ ctrl.LED, tciClient *client.Client) (interface{}, ControllerType, error) {
// 		return ctrl.NewVolumeSlider(tciClient), SliderController, nil
// 	},
// 	BalanceMapping: func(m Mapping, _ ctrl.LED, tciClient *client.Client) (interface{}, ControllerType, error) {
// 		vfo, err := AtoVFO(m.VFO)
// 		if err != nil {
// 			return nil, 0, err
// 		}
// 		return ctrl.NewRXBalanceSlider(m.TRX, vfo, tciClient), SliderController, nil
// 	},
// 	EnableRXMapping:         nil,
// 	EnableSplitMapping:      nil,
// 	EnableRITMapping:        nil,
// 	RITMapping:              nil,
// 	EnableXITMapping:        nil,
// 	XITMapping:              nil,
// 	SyncVFOFrequencyMapping: nil,
// 	FilterMapping:           nil,
// 	ModeMapping:             nil,
// }
