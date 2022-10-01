package ctrl

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/ftl/tci/client"
)

const VFOMapping MappingType = "vfo"

func init() {
	Factories[VFOMapping] = func(m Mapping, _ LED, tciClient *client.Client) (interface{}, ControlType, error) {
		vfo, err := AtoVFO(m.VFO)
		if err != nil {
			return nil, 0, err
		}
		reverseDirection := false
		directionStr := m.Options["direction"]
		if strings.ToLower(directionStr) == "reverse" {
			reverseDirection = true
		}
		dynamicMode := false
		modeStr := m.Options["mode"]
		if strings.ToLower(modeStr) == "dynamic" {
			dynamicMode = true
		}
		stepSizeStr, ok := m.Options["step"]
		var stepSize int
		if ok {
			stepSize, err = strconv.Atoi(stepSizeStr)
			if err != nil {
				return nil, ButtonControl, fmt.Errorf("the step size is invalid: %v", err)
			}
		}
		if stepSize == 0 {
			stepSize = 10
		}

		return NewVFOEncoder(m.TRX, vfo, stepSize, reverseDirection, dynamicMode, tciClient), EncoderControl, nil
	}
}

func NewVFOEncoder(trx int, vfo client.VFO, stepSize int, reverseDirection bool, dynamicMode bool, controller VFOFrequencyController) *VFOEncoder {
	return &VFOEncoder{
		Encoder: NewEncoder(
			func(frequency int) {
				err := controller.SetVFOFrequency(trx, vfo, frequency)
				if err != nil {
					log.Printf("Cannot change frequency to %d: %v", frequency, err)
				}
			},
			InfiniteRange{},
			stepSize,
			reverseDirection,
			dynamicMode,
		),
		trx: trx,
		vfo: vfo,
	}
}

type VFOEncoder struct {
	*Encoder
	trx int
	vfo client.VFO
}

func (e *VFOEncoder) SetVFOFrequency(trx int, vfo client.VFO, frequency int) {
	if trx != e.trx || vfo != e.vfo {
		return
	}
	e.Encoder.SetActiveValue(frequency)
}

type VFOFrequencyController interface {
	SetVFOFrequency(trx int, vfo client.VFO, frequency int) error
}
