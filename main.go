package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type WorkerResult struct {
	Err      error
	Time     time.Duration
	Status   int
	WorkerID int
}

func main() {
	// TODO: consider othre requests than GET
	workers := flag.Int("w", 1, "number of concurrent workers")
	requests := flag.Int("r", 1, "number of requests per worker")
	verbose := flag.Bool("v", false, "verbose output (errors)")
	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Fprintln(os.Stderr, "missing URL")
		os.Exit(1)
	}
	if *workers < 1 {
		fmt.Fprintln(os.Stderr, "must use at least one worker")
		os.Exit(1)
	}
	if *requests < 1 {
		fmt.Fprintln(os.Stderr, "must perform at least one request")
		os.Exit(1)
	}
	url := flag.Args()[0]
	results := perform(url, *workers, *requests, *verbose)
	fmt.Println(overallMean(results))
}

func overallMean(results map[int][]WorkerResult) time.Duration {
	var duration time.Duration
	var i int
	for _, rs := range results {
		for _, r := range rs {
			duration += r.Time
			i++
		}
	}
	return duration / time.Duration(i)
}

func perform(url string, workers, requests int, verbose bool) map[int][]WorkerResult {
	var wg sync.WaitGroup
	overall := make(map[int][]WorkerResult)
	overallChan := make(chan map[int][]WorkerResult)
	resultChan := make(chan WorkerResult)
	go func() {
		for result := range resultChan {
			if existing, ok := overall[result.WorkerID]; ok {
				overall[result.WorkerID] = append(existing, result)
			} else {
				workerResults := make([]WorkerResult, 1)
				workerResults[0] = result
				overall[result.WorkerID] = workerResults
			}
		}
		overallChan <- overall
	}()
	for w := 0; w < workers; w++ {
		for r := 0; r < requests; r++ {
			wg.Add(1)
			go func(ch chan WorkerResult, id int) {
				ch <- request(url, id, verbose)
				wg.Done()
			}(resultChan, w)
		}
	}
	wg.Wait()
	close(resultChan)
	return <-overallChan
}

func request(url string, workerID int, verbose bool) WorkerResult {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "create request for %s: %v", url, err)
		}
		return WorkerResult{err, 0.0, 0, workerID}
	}
	start := time.Now()
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "perform request for %s: %v", url, err)
		}
		return WorkerResult{err, time.Since(start), 0, workerID}
	}
	defer res.Body.Close()
	return WorkerResult{nil, time.Since(start), res.StatusCode, workerID}
}
