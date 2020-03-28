package main

import (
	"context"
	"sync"

	workerpool "github.com/vardius/worker-pool/v2"
)

// Handler function
type Handler interface{}

// Payload of the message
type Payload []byte

// Service allows to push/pull messages
type Service interface {
	Delegate(ctx context.Context, topic string, p Payload) error
	AddWorker(topic string, fn Handler) error
	RemoveWorker(topic string, fn Handler) error
}

type pools map[string]workerpool.Pool

type service struct {
	maxConcurrentCalls int
	pools              pools
	mtx                sync.RWMutex
}

// newService creates in memory command pool
func newService(maxConcurrentCalls int) Service {
	return &service{
		maxConcurrentCalls: maxConcurrentCalls,
		pools:              make(pools),
	}
}

func (s *service) Delegate(ctx context.Context, topic string, p Payload) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, ok := s.pools[topic]; !ok {
		return nil
	}

	return s.pools[topic].Delegate(ctx, p)
}

func (s *service) AddWorker(topic string, fn Handler) error {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	if _, ok := s.pools[topic]; !ok {
		s.pools[topic] = workerpool.New(s.maxConcurrentCalls)
	}

	if err := s.pools[topic].AddWorker(fn); err != nil {
		return err
	}

	return nil
}

func (s *service) RemoveWorker(topic string, fn Handler) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, ok := s.pools[topic]; !ok {
		return nil
	}

	if err := s.pools[topic].RemoveWorker(fn); err != nil {
		return err
	}

	if s.pools[topic].WorkersNum() == 0 {
		s.pools[topic].Stop()

		delete(s.pools, topic)
	}

	return nil
}
