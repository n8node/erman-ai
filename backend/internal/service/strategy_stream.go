package service

import (
	"encoding/json"
	"sync"
)

type StrategyStreamEventType string

const (
	StrategyEventPhase  StrategyStreamEventType = "phase"
	StrategyEventChunk  StrategyStreamEventType = "chunk"
	StrategyEventDone   StrategyStreamEventType = "done"
	StrategyEventRunError StrategyStreamEventType = "run_error"
)

type StrategyStreamEvent struct {
	Type StrategyStreamEventType `json:"type"`
	Data json.RawMessage         `json:"data"`
}

type strategyRunStream struct {
	mu      sync.Mutex
	events  []StrategyStreamEvent
	subs    map[chan StrategyStreamEvent]struct{}
	closed  bool
}

type StrategyStreamHub struct {
	mu   sync.RWMutex
	runs map[string]*strategyRunStream
}

func NewStrategyStreamHub() *StrategyStreamHub {
	return &StrategyStreamHub{runs: make(map[string]*strategyRunStream)}
}

func (h *StrategyStreamHub) ensure(runID string) *strategyRunStream {
	h.mu.Lock()
	defer h.mu.Unlock()
	if s, ok := h.runs[runID]; ok {
		return s
	}
	s := &strategyRunStream{subs: make(map[chan StrategyStreamEvent]struct{})}
	h.runs[runID] = s
	return s
}

func (h *StrategyStreamHub) Publish(runID string, ev StrategyStreamEvent) {
	s := h.ensure(runID)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, ev)
	for ch := range s.subs {
		select {
		case ch <- ev:
		default:
		}
	}
	if ev.Type == StrategyEventDone || ev.Type == StrategyEventRunError {
		s.closed = true
	}
}

func (h *StrategyStreamHub) Subscribe(runID string) (chan StrategyStreamEvent, []StrategyStreamEvent, bool) {
	s := h.ensure(runID)
	s.mu.Lock()
	defer s.mu.Unlock()
	replay := append([]StrategyStreamEvent(nil), s.events...)
	ch := make(chan StrategyStreamEvent, 64)
	if s.closed {
		return ch, replay, true
	}
	s.subs[ch] = struct{}{}
	return ch, replay, s.closed
}

func (h *StrategyStreamHub) Unsubscribe(runID string, ch chan StrategyStreamEvent) {
	h.mu.RLock()
	s, ok := h.runs[runID]
	h.mu.RUnlock()
	if !ok {
		return
	}
	s.mu.Lock()
	delete(s.subs, ch)
	s.mu.Unlock()
}

func (h *StrategyStreamHub) Cleanup(runID string) {
	h.mu.Lock()
	delete(h.runs, runID)
	h.mu.Unlock()
}

func strategyEventData(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
