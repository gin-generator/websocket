package websocket

import (
	"errors"
	"sync"
)

// Memory is the subscription store: it maps channels to subscriber connection ids.
type Memory interface {
	Set(id, channel string) error
	GetSubscribers(channel string) (ids []string, err error)
	Delete(id, channel string) error
}

type SystemMemory struct {
	subscribe map[string][]string
	mux       *sync.RWMutex
}

func newSystemMemory() *SystemMemory {
	return &SystemMemory{
		subscribe: make(map[string][]string, RateLimit),
		mux:       new(sync.RWMutex),
	}
}

func (s *SystemMemory) Set(id, channel string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if _, ok := s.subscribe[channel]; !ok {
		s.subscribe[channel] = make([]string, RateLimit)
	}
	for _, v := range s.subscribe[channel] {
		if v == id {
			return nil
		}
	}
	s.subscribe[channel] = append(s.subscribe[channel], id)
	return nil
}

func (s *SystemMemory) GetSubscribers(channel string) (ids []string, err error) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	if v, ok := s.subscribe[channel]; ok {
		return v, nil
	}
	return nil, errors.New("channel not found")
}

func (s *SystemMemory) Delete(id, channel string) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if v, ok := s.subscribe[channel]; ok {
		for i, vv := range v {
			if vv == id {
				s.subscribe[channel] = append(v[:i], v[i+1:]...)
				return nil
			}
		}
		return nil
	}
	return errors.New("channel not found")
}
