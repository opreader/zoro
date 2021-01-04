package spinner

import (
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/fatih/color"
)

type Spinner struct {
	mu         *sync.RWMutex
	Delay      time.Duration                 // Delay is the speed of the indicator
	writer     io.Writer                     // Use WithWriter() to update after initialization
	chars      []string                      // chars holds the chosen character set
	color      func(a ...interface{}) string // default color is white
	stopChan   chan struct{}                 // stopChan is a channel used to stop the indicator
	lastOutput string                        // last character(set) written
	active     bool                          // active holds the state of the spinner
	PreUpdate  func(s *Spinner)
	PostUpdate func(s *Spinner)
	Prefix     string
	Suffix     string
	FinalMSG   string
}

func New(cs []string, d time.Duration, opts ...Option) *Spinner {
	s := &Spinner{
		Delay:    d,
		chars:    cs,
		writer:   color.Output,
		color:    color.New(color.FgWhite).SprintFunc(),
		mu:       &sync.RWMutex{},
		stopChan: make(chan struct{}, 1),
		active:   false,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Spinner) Start() {
	s.mu.Lock()
	if s.active {
		s.mu.Unlock()
		return
	}
	s.active = true
	s.mu.Unlock()

	go func() {
		for {
			for i := 0; i < len(s.chars); i++ {
				select {
				case <-s.stopChan:
					return
				default:
					if !s.active {
						return
					}
					s.mu.Lock()
					s.erase()

					if s.PreUpdate != nil {
						s.PreUpdate(s)
					}

					outColor := fmt.Sprintf("%s%s%s ", s.Prefix, s.color(s.chars[i]), s.Suffix)
					outPlain := fmt.Sprintf("%s%s%s ", s.Prefix, s.chars[i], s.Suffix)
					fmt.Fprint(s.writer, outColor)
					s.lastOutput = outPlain
					delay := s.Delay

					if s.PostUpdate != nil {
						s.PostUpdate(s)
					}

					s.mu.Unlock()
					time.Sleep(delay)
				}
			}
		}
	}()
}

func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active {
		s.active = false
		s.erase()
		if len(s.FinalMSG) > 0 {
			fmt.Fprintf(s.writer, s.FinalMSG)
		}
		s.stopChan <- struct{}{}
	}
}

func (s *Spinner) Restart() {
	s.Stop()
	s.Start()
}

// set the struct field for the given color to be used.
func (s *Spinner) Color(colors ...color.Attribute) error {
	s.mu.Lock()
	s.color = color.New(colors...).SprintFunc()
	s.mu.Unlock()
	s.Restart()
	return nil
}

// set the indicator delay to the given value.
func (s *Spinner) UpdateSpeed(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Delay = d
}

// change the current character set to the given one.
func (s *Spinner) UpdateCharSet(cs []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.chars = cs
}

func (s *Spinner) Active() bool {
	return s.active
}

// deletes written characters. Caller must already hold s.lock.
func (s *Spinner) erase() {
	n := utf8.RuneCountInString(s.lastOutput)
	del, _ := hex.DecodeString("7f")
	for _, c := range []string{"\b", string(del)} {
		for i := 0; i < n; i++ {
			fmt.Fprintf(s.writer, c)
		}
	}
	fmt.Fprintf(s.writer, "\r\033[K") // erases to end of line
	s.lastOutput = ""
}

// a function that takes a spinner and applies a given configuration.
type Option func(*Spinner)

// adds the given writer to the spinner. This function should be favored over directly assigning to the struct value.
func WithWriter(w io.Writer) Option {
	return func(s *Spinner) {
		s.mu.Lock()
		s.writer = w
		s.mu.Unlock()
	}
}

func WithColor(color color.Attribute) Option {
	return func(s *Spinner) {
		s.Color(color)
	}
}

func WithSuffix(suffix string) Option {
	return func(s *Spinner) {
		s.Suffix = suffix
	}
}

func WithFinalMSG(finalMsg string) Option {
	return func(s *Spinner) {
		s.FinalMSG = finalMsg
	}
}
