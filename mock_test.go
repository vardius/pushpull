package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/empty"

	pushpull_mock "github.com/vardius/pushpull/mock_proto"
	pushpull_proto "github.com/vardius/pushpull/proto"
)

var topic = "my-topic"
var msg = []byte("Hello you!")

var emptyResponse *empty.Empty

var pullResponse = &pushpull_proto.PullResponse{Payload: msg}
var pushRequest = &pushpull_proto.PushRequest{Topic: topic, Payload: msg}
var pullRequest = &pushpull_proto.PullRequest{Topic: topic}

func TestServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock for the stream returned by Pull
	stream := pushpull_mock.NewMockPushPull_PullClient(ctrl)
	// Set expectation on receiving.
	stream.EXPECT().Recv().Return(pullResponse, nil)
	stream.EXPECT().CloseSend().Return(nil)

	// Create mock for the client interface.
	client := pushpull_mock.NewMockPushPullClient(ctrl)
	// Set expectation on Push
	client.EXPECT().Push(gomock.Any(), pushRequest).Return(emptyResponse, nil)
	// Set expectation on Pull
	client.EXPECT().Pull(gomock.Any(), pullRequest).Return(stream, nil)

	if err := testPushPull(client); err != nil {
		t.Fatalf("Test failed: %v", err)
	}
}

func testPushPull(client pushpull_proto.PushPullClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// test Pull
	stream, err := client.Pull(ctx, pullRequest)
	if err != nil {
		return err
	}
	if err := stream.CloseSend(); err != nil {
		return err
	}
	got, err := stream.Recv()
	if err != nil {
		return err
	}
	if !proto.Equal(got, pullResponse) {
		return fmt.Errorf("stream.Recv() = %v, want %v", got, pullResponse)
	}

	// test Push
	_, err = client.Push(ctx, pushRequest)
	if err != nil {
		return err
	}

	return nil
}
