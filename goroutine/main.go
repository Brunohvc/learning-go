package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

func main() {
	// sequenceRun()
	// cocurrenceRun()
	contextRun()
}

func sequenceRun() {
	start := time.Now()
	for i := 0; i < 10; i++ {
		resp, err := http.Get("https://www.google.com")
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		fmt.Println(resp.Status, "OK", i)
	}
	fmt.Println(time.Since(start))
}

func cocurrenceRun() {
	start := time.Now()
	interactions := 10

	wg := sync.WaitGroup{}
	wg.Add(interactions)
	for i := 0; i < interactions; i++ {
		go func() {
			defer wg.Done()
			resp, err := http.Get("https://www.google.com")
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			fmt.Println(resp.Status, "OK", i)
		}()
	}
	wg.Wait()
	fmt.Println(time.Since(start))
}

func contextRun() {
	start := time.Now()
	interactions := 10

	wg := sync.WaitGroup{}
	wg.Add(interactions)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second)
		fmt.Fprintln(w, "Hello, client")
	}))

	for range interactions {
		go func(ctx context.Context) {
			defer wg.Done()

			req, err := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
			if err != nil {
				panic(err)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					fmt.Println("Request canceled")
					return
				}
				panic(err)
			}

			defer resp.Body.Close()
		}(ctx)
	}

	wg.Wait()

	fmt.Println(time.Since(start))
}
