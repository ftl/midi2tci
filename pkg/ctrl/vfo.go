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
		direction := 1
		directionStr := m.Options["direction"]
		if strings.ToLower(directionStr) == "reverse" {
			direction = -1
		}
		stepSizeStr, ok := m.Options["step"]
		stepSize := 10
		if ok {
			stepSize, err = strconv.Atoi(stepSizeStr)
			if err != nil {
				return nil, ButtonController, fmt.Errorf("the step size is invalid: %v", err)
			}
		}

		return NewVFOWheel(m.MidiKey(), m.TRX, vfo, direction, stepSize, tciClient), WheelController, nil
	}
}

func NewVFOWheel(key MidiKey, trx int, vfo client.VFO, direction int, stepSize int, controller VFOFrequencyController) *VFOWheel {
	result := &VFOWheel{
		key:        key,
		trx:        trx,
		vfo:        vfo,
		controller: controller,
		frequency:  make(chan int, 1000),
		turns:      make(chan int, 1000),
		closed:     make(chan struct{}),
	}

	go func() {
		const (
			scanInterval = 10 * time.Millisecond
			velocity     = 1
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
				if stepSize > 0 {
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
