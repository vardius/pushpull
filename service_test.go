package main

import (
	"context"
	"runtime"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	s := newService(runtime.NumCPU())

	if s == nil {
		t.Fail()
	}
}

func TestDelegate(t *testing.T) {
	s := newService(runtime.NumCPU())
	ctx := context.Background()
	c := make(chan error)

	worker := func(ctx context.Context, _ []byte) {
		c <- nil
	}

	if err := s.AddWorker("topic", worker); err != nil {
		t.Fatal(err)
	}

	if err := s.Delegate(ctx, "topic", []byte("ok")); err != nil {
		t.Fatal(err)
	}

	ctxDoneCh := ctx.Done()
	for {
		select {
		case <-ctxDoneCh:
			t.Fatal(ctx.Err())
			return
		case err := <-c:
			if err != nil {
				t.Error(err)
			}
			return
		}
	}
}

func TestRemoveWorker(t *testing.T) {
	s := newService(runtime.NumCPU())
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	handler := func(ctx context.Context, _ []byte) {
		t.Fail()
	}

	if err := s.AddWorker("topic", handler); err != nil {
		t.Fatal(err)
	}
	if err := s.RemoveWorker("topic", handler); err != nil {
		t.Fatal(err)
	}
	if err := s.Delegate(ctx, "topic", []byte("ok")); err != nil {
		t.Fatal(err)
	}

	ctxDoneCh := ctx.Done()
	for {
		select {
		case <-ctxDoneCh:
			return
		}
	}
}
