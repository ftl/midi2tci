package ctrl

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/ftl/tci/client"
)

const VFOMapping MappingType = "vfo"

func init() {
	Factories[VFOMapping] = func(m Mapping, _ LED, tciClient *client.Client) (interface{}, ControllerType, error) {
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
				return nil, ButtonController, fmt.Errorf("the step size is invalid: %v", err)
			}
		}
		if stepSize == 0 {
			stepSize = 10
		}

		return NewVFOWheel(m.MidiKey(), m.TRX, vfo, reverseDirection, stepSize, dynamicMode, tciClient), WheelController, nil
	}
}

func NewVFOWheel(key MidiKey, trx int, vfo client.VFO, reverseDirection bool, stepSize int, dynamicMode bool, controller VFOFrequencyController) *VFOWheel {
	result := &VFOWheel{
		key:        key,
		trx:        trx,
		vfo:        vfo,
		controller: controller,
		frequency:  make(chan int, 1000),
		turns:      make(chan int, 1000),
		closed:     make(chan struct{}),
	}
	direction := 1
	if reverseDirection {
		direction = -1
	}

	go func() {
		const (
			scanInterval = 10 * time.Millisecond
		)

		defer close(result.closed)
		ticker := time.NewTicker(scanInterval)
		defer ticker.Stop()

		accumulatedTurns := 0
		turning := false
		frequency := 0
		for {
			select {
			case turns, valid := <-result.turns:
				if !valid {
					return
				}
				if dynamicMode {
					turns = stepSize * turns
				} else {
					if turns < 0 {
						turns = -stepSize
					} else if turns > 0 {
						turns = stepSize
					}
				}
				accumulatedTurns += (turns * direction)
				turning = frequency > 0
			case f := <-result.frequency:
				if !turning {
					frequency = f
				}
			case <-ticker.C:

				if accumulatedTurns == 0 {
					turning = false
					break
				}
				if frequency == 0 {
					break
				}

				nextFrequency := int(math.Round(float64(frequency+accumulatedTurns)/float64(stepSize))) * stepSize
				usedSteps := nextFrequency - frequency

				frequency = nextFrequency
				accumulatedTurns -= usedSteps
				if accumulatedTurns < 0 {
					accumulatedTurns = 0
				}

				err := result.controller.SetVFOFrequency(result.trx, result.vfo, frequency)
				if err != nil {
					log.Printf("Cannot change frequency to %d: %v", result.frequency, err)
				}
			}
		}
	}()

	return result
}

type VFOWheel struct {
	key        MidiKey
	trx        int
	vfo        client.VFO
	controller VFOFrequencyController

	frequency chan int
	turns     chan int
	closed    chan struct{}
}

type VFOFrequencyController interface {
	SetVFOFrequency(trx int, vfo client.VFO, frequency int) error
}

func (w *VFOWheel) Close() {
	select {
	case <-w.closed:
		return
	default:
		close(w.turns)
		<-w.closed
	}
}

func (w *VFOWheel) Turned(turns int) {
	w.turns <- turns
}

func (w *VFOWheel) SetVFOFrequency(trx int, vfo client.VFO, frequency int) {
	if trx != w.trx || vfo != w.vfo {
		return
	}
	w.frequency <- frequency
}
