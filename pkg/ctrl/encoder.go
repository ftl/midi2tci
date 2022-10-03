package ctrl

import (
	"math"
	"time"
)

func NewEncoder(key MidiKey, set func(int), valueRange ValueRange, led LED, stepSize int, reverseDirection bool, dynamicMode bool) *Encoder {
	result := &Encoder{
		key:         key,
		set:         set,
		valueRange:  valueRange,
		led:         led,
		activeValue: make(chan int, 1000),
		turns:       make(chan int, 1000),
		closed:      make(chan struct{}),

		stepSize:         stepSize,
		reverseDirection: reverseDirection,
		dynamicMode:      dynamicMode,
	}

	result.start()

	return result
}

type Encoder struct {
	key         MidiKey
	set         func(int)
	valueRange  ValueRange
	led         LED
	activeValue chan int
	turns       chan int
	closed      chan struct{}

	stepSize         int
	reverseDirection bool
	dynamicMode      bool
}

func (e *Encoder) start() {
	direction := 1
	if e.reverseDirection {
		direction = -1
	}

	tx := make(chan int)
	go func() {
		for {
			select {
			case <-e.closed:
				return
			case value := <-tx:
				e.set(TrimToRange(e.valueRange, value))
			}
		}
	}()

	go func() {
		defer close(e.closed)
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		activeValue := 0
		selectedValue := 0
		accumulatedTurns := 0
		pending := false

		for {
			select {
			case value, valid := <-e.activeValue:
				if !valid {
					return
				}
				activeValue = value
				// log.Printf("encoder active value: %d", activeValue)
				if !pending {
					selectedValue = activeValue
				}
			case turns, valid := <-e.turns:
				if !valid {
					return
				}
				// log.Printf("1 turns: %d", turns)

				if e.dynamicMode {
					turns = e.stepSize * turns
				} else {
					if turns < 0 {
						turns = -e.stepSize
					} else if turns > 0 {
						turns = e.stepSize
					}
				}
				// log.Printf("2 turns: %d", turns)

				accumulatedTurns += (turns * direction)
				if accumulatedTurns == 0 {
					pending = false
					continue
				}

				nextValue := int(math.Round(float64(activeValue+accumulatedTurns)/float64(e.stepSize))) * e.stepSize
				nextValue = TrimToRange(e.valueRange, nextValue)
				usedSteps := nextValue - activeValue

				if usedSteps == 0 {
					accumulatedTurns = 0
					pending = false
					continue
				}

				selectedValue = nextValue
				accumulatedTurns -= usedSteps
				if activeValue == selectedValue {
					continue
				}

				select {
				case tx <- selectedValue:
					activeValue = selectedValue
					pending = false
				default:
					pending = true
				}
			case <-ticker.C:
				if activeValue == selectedValue {
					pending = false
					continue
				}

				select {
				case tx <- selectedValue:
					activeValue = selectedValue
					pending = false
				default:
					pending = true
				}
			}
		}
	}()
}

func (e *Encoder) Close() {
	select {
	case <-e.closed:
		return
	default:
		close(e.activeValue)
		close(e.turns)
		<-e.closed
	}
}

func (e *Encoder) Changed(turns int) {
	e.turns <- turns
}

func (e *Encoder) SetActiveValue(value int) {
	e.activeValue <- value
	if e.led != nil {
		e.led.SetValue(e.key, Project(e.valueRange, value))
	}
}
