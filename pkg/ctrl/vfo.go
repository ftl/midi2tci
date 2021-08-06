package ctrl

import (
	"log"
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
		return NewVFOWheel(m.MidiKey(), m.TRX, vfo, tciClient), WheelController, nil
	}
}

func NewVFOWheel(key MidiKey, trx int, vfo client.VFO, controller VFOFrequencyController) *VFOWheel {
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
		defer close(result.closed)
		ticker := time.NewTicker(10 * time.Millisecond)
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
				accumulatedTurns += turns
				turning = frequency > 0
			case f := <-result.frequency:
				if !turning {
					frequency = f
				}
			case <-ticker.C:
				if accumulatedTurns == 0 {
					turning = false
				} else if accumulatedTurns != 0 && frequency != 0 {
					frequency = frequency + int(float64(accumulatedTurns)*1.8)
					err := result.controller.SetVFOFrequency(result.trx, result.vfo, frequency)
					if err != nil {
						log.Printf("Cannot change frequency to %d: %v", result.frequency, err)
					}
					accumulatedTurns = 0
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
