package ctrl

import "time"

func NewSlider(set func(int), translate func(int) int) *Slider {
	result := &Slider{
		set:           set,
		translate:     translate,
		selectedValue: make(chan int, 1000),
		activeValue:   make(chan int, 1000),
		closed:        make(chan struct{}),
	}

	result.start()

	return result
}

type Slider struct {
	set           func(int)
	translate     func(int) int
	activeValue   chan int
	selectedValue chan int
	closed        chan struct{}
}

func (s *Slider) start() {
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
				if !pending {
					selectedValue = activeValue
				}
			case value, valid := <-s.selectedValue:
				if !valid {
					return
				}
				selectedValue = value

				if activeValue == selectedValue {
					continue
				}

				select {
				case tx <- selectedValue:
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
					pending = false
				default:
					pending = true
				}
			}
		}
	}()
}

func (s *Slider) Close() {
	select {
	case <-s.closed:
		return
	default:
		close(s.activeValue)
		close(s.selectedValue)
		<-s.closed
	}
}

func (s *Slider) Changed(value int) {
	s.selectedValue <- s.translate(value)
}

func (s *Slider) SetActiveValue(value int) {
	s.activeValue <- value
}
