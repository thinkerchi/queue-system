package main

import (
	"sync"
	"testing"
)

func TestClient(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	for i := 0; i < 10000; i++ {
		go run()
	}
	wg.Wait()
}

func BenchmarckClient(b *testing.B) {

}
