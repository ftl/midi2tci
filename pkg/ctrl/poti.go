package ctrl

import (
	"time"
)

func NewPoti(key MidiKey, set func(int), valueRange ValueRange, led LED) *Poti {
	result := &Poti{
		key:           key,
		set:           set,
		valueRange:    valueRange,
		led:           led,
		selectedValue: make(chan int, 1000),
		activeValue:   make(chan int, 1000),
		closed:        make(chan struct{}),
	}

	result.start()

	return result
}

type Poti struct {
	key           MidiKey
	set           func(int)
	valueRange    ValueRange
	led           LED
	activeValue   chan int
	selectedValue chan int
	closed        chan struct{}
}

func (s *Poti) start() {
	tx := make(chan int)
	go func() {
		for {
			select {
			case <-s.closed:
				return
			case value := <-tx:
				s.set(value)
			}
		}
	}()

	go func() {
		defer close(s.closed)
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		activeValue := 0
		selectedValue := 0
		pending := false

		for {
			select {
			case value, valid := <-s.activeValue:
				if !valid {
					return
				}
				activeValue = value
				// log.Printf("poti active value: %d", activeValue)
				if !pending {
					selectedValue = activeValue
				}
			case value, valid := <-s.selectedValue:
				if !valid {
					return
				}
				selectedValue = value
				// log.Printf("poti selectedValue: %d", selectedValue)

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

func (s *Poti) Close() {
	select {
	case <-s.closed:
		return
	default:
		close(s.activeValue)
		close(s.selectedValue)
		<-s.closed
	}
}

func (s *Poti) Changed(value int) {
	s.selectedValue <- Translate(s.valueRange, uint8(value))
}

func (s *Poti) SetActiveValue(value int) {
	s.activeValue <- value
	if s.led != nil {
		s.led.SetValue(s.key, Project(s.valueRange, value))
	}
}
