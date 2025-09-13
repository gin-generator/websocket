package websocket

import (
	"errors"
)

type Memory interface {
	Set(id, channel string) error
	Get(id string) (ids []string, err error)
	Delete(id, channel string) error
}

type SystemMemory struct {
	subscribe map[string][]string
}

func newSystemMemory() *SystemMemory {
	return &SystemMemory{
		subscribe: make(map[string][]string, RateLimit),
	}
}

func (s *SystemMemory) Set(id, channel string) error {
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

func (s *SystemMemory) Get(channel string) (ids []string, err error) {
	if v, ok := s.subscribe[channel]; ok {
		return v, nil
	}
	return nil, errors.New("channel not found")
}

func (s *SystemMemory) Delete(id, channel string) error {
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
