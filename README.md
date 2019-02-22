# Simple in-memory pubsub library

## Requirements
 * [Golang 1.11+](https://golang.org/dl/)
 * [dep 0.5.0](https://github.com/golang/dep#installation)
 * [golangci-lint](https://github.com/golangci/golangci-lint)  
or 
 * [nix](https://nixos.org/nix/download.html) - just run `nix-shell` in project's folder and it will drop you into development environment

## Usage
```go
package main

import (
	"fmt"
	"log"
)

func main() {
	s := pubsub.NewServer()
	topic := "topic"
	subscriber := "sub"
	msg := []byte(`{"foo":"bar"}`)
	s.Subscribe(topic, subscriber)
	s.Publish(topic, msg)
	msg, err := s.Poll(topic, subscriber)
	if err != nil {
		log.Fatalf("Unexpected error: %v", err)
	}
	s.Unsubscribe(topic, subscriber)

	fmt.Println(string(msg))
	// Output: {"foo":"bar"}
}

```

## Testing
Run:
```sh
golangci-ling run && go test -v -timeout=30s --cover --race ./...
```
or with nix: 
```sh
nix-shell --run "pubsub-lint-run-tests"
```

### Ideas to test
Current implementation uses mutexes for concurrency control and linked list from stdlib to store messages.
It would be great to implement and compare performance with next implementations:
 * channel based communication
 * [stm](https://github.com/lukechampine/stm) instead of mutexes
 * Review possible implementations for concurrent queues (right now we have RWMutex on linked list for all operations)
