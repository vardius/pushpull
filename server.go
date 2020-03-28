package main

import (
	"context"
	"io"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/vardius/golog"

	"github.com/vardius/pushpull/proto"
)

type server struct {
	service Service
	logger  golog.Logger
}

// newServer returns new push/pull server object
func newServer(service Service, logger golog.Logger) proto.PushPullServer {
	return &server{service, logger}
}

// Push pushed message to the worker queue
func (s *server) Push(ctx context.Context, r *proto.PushRequest) (*empty.Empty, error) {
	s.logger.Debug(ctx, "Push: %s %s", r.GetTopic(), r.GetPayload())

	if err := s.service.Delegate(ctx, r.GetTopic(), r.GetPayload()); err != nil {
		return nil, err
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return new(empty.Empty), nil
}

// Pull pulls message from the queue,
// will stop pulling messages when stream.Send returns error
func (s *server) Pull(r *proto.PullRequest, stream proto.PushPull_PullServer) error {
	errCh := make(chan error)
	defer close(errCh)

	handler := func(ctx context.Context, payload Payload) {
		s.logger.Debug(ctx, "Pull: %s %s", r.GetTopic(), payload)

		err := stream.Send(&proto.PullResponse{
			Payload: payload,
		})

		if err != nil {
			errCh <- err
		}
	}

	ctx := context.Background()

	s.logger.Info(ctx, "AddWorker: %s", r.GetTopic())

	if err := s.service.AddWorker(r.GetTopic(), handler); err != nil {
		return err
	}

	err := <-errCh

	if err := s.service.RemoveWorker(r.GetTopic(), handler); err != nil {
		return err
	}

	if err == io.EOF {
		s.logger.Info(ctx, "RemoveWorker: %s - Stream closed, no more input is available", r.GetTopic())

		return nil
	}

	s.logger.Info(ctx, "RemoveWorker: %s - %s", r.GetTopic(), err.Error())

	return err
}
