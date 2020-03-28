# pushpull

[![](https://godoc.org/github.com/vardius/pushpull/proto?status.svg)](https://pkg.go.dev/github.com/vardius/pushpull/proto)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg)](https://github.com/vardius/pushpull/blob/master/LICENSE.md)

Package proto contains protocol buffer code to populate

<details>
  <summary>Table of Content</summary>

<!-- toc -->
- [How to use](#how-to-use)
  - [Client](https://github.com/vardius/pushpull/tree/master/proto#client)
  	- [Use in your Go project](https://github.com/vardius/pushpull/tree/master/proto#use-in-your-go-project)
	  - [Push](https://github.com/vardius/pushpull/tree/master/proto#push)
	  - [Pull](https://github.com/vardius/pushpull/tree/master/proto#pull)
  - [Protocol Buffers](https://github.com/vardius/pushpull/tree/master/proto#protocol-buffers)
	- [Generating client and server code](https://github.com/vardius/pushpull/tree/master/proto#generating-client-and-server-code)
<!-- tocstop -->
</details>

# HOW TO USE

## Client

### Use in your Go project

#### Push

```go
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"github.com/vardius/pushpull/proto"
)

func main() {
    host:= "0.0.0.0"
    port:= 9090
    ctx := context.Background()

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
			Timeout:             20 * time.Second, // wait 20 second for ping ack before considering the connection dead
			PermitWithoutStream: true,             // send pings even without active streams
		}),
    }

	conn, err := grpc.DialContext(ctx, fmt.Sprintf("%s:%d", host, port), opts...)
	if err != nil {
		os.Exit(1)
    }
    defer conn.Close()

	client := proto.NewPushPullClient(conn)

    client.Push(ctx, &proto.PushRequest{
		Topic:   "my-topic",
		Payload: []byte("Hello you!"),
    })
}
```

#### Pull

```go
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"github.com/vardius/pushpull/proto"
)

func main() {
    host:= "0.0.0.0"
    port:= 9090
    ctx := context.Background()

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
			Timeout:             20 * time.Second, // wait 20 second for ping ack before considering the connection dead
			PermitWithoutStream: true,             // send pings even without active streams
		}),
    }

	conn, err := grpc.DialContext(ctx, fmt.Sprintf("%s:%d", host, port), opts...)
	if err != nil {
		os.Exit(1)
    }
    defer conn.Close()

	client := proto.NewPushPullClient(conn)

	stream, err := client.Pull(ctx, &proto.PullRequest{
		Topic: "my-topic",
	})
	if err != nil {
		os.Exit(1)
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
		    os.Exit(1) // stream closed or error
		}

		fmt.Println(resp.GetPayload())
	}
}
```

## Protocol Buffers

### Generating client and server code

To generate the gRPC client and server interfaces from `pushpull.proto` service definition.
Use the protocol buffer compiler protoc with a special gRPC Go plugin. For more info [read](https://grpc.io/docs/quickstart/go.html)

From this directory run:

```bash
$ make build
```

Running this command generates the following files in this directory:

- `pushpull.pb.go`

This contains:

All the protocol buffer code to populate, serialize, and retrieve our request and response message types
An interface type (or stub) for clients to call with the methods defined in the services.
An interface type for servers to implement, also with the methods defined in the services.
