package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
)

type WorkerResult struct {
	OK       bool
	Time     time.Duration
	WorkerID int
}

type Results map[int][]WorkerResult

type Stats struct {
	Total  int
	Passed int
	Failed int
	Mean   time.Duration
	Perc25 time.Duration
	Perc50 time.Duration
	Perc75 time.Duration
}

func main() {
	workers := flag.Int("w", 1, "number of concurrent workers")
	requests := flag.Int("r", 1, "number of requests per worker")
	state := flag.Int("s", http.StatusOK, "response state considered success")
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

	results := run(flag.Args()[0], *workers, *requests, *state)

	s := stats(results)
	nTotal := *workers * *requests
	fmt.Printf("%12s %12s %12s %12s %10s %10s %10s\n",
		"mean", "25%", "50%", "75%", "requests", "passed", "failed")
	fmt.Printf("%12s %12s %12s %12s %10d %10d %10d\n",
		s.Mean, s.Perc25, s.Perc50, s.Perc75, nTotal, s.Passed, s.Failed)
}

func run(url string, workers, requests, okState int) Results {
	var wg sync.WaitGroup
	overall := make(chan Results)
	results := make(chan WorkerResult)

	go collect(overall, results)

	for w := 0; w < workers; w++ {
		for r := 0; r < requests; r++ {
			wg.Add(1)
			go func(ch chan WorkerResult, id int) {
				ch <- get(url, okState, w)
				wg.Done()
			}(results, w)
		}
	}
	wg.Wait()
	close(results)

	return <-overall
}

func collect(whole chan<- Results, parts <-chan WorkerResult) {
	overall := make(Results)
	for result := range parts {
		if existing, ok := overall[result.WorkerID]; ok {
			overall[result.WorkerID] = append(existing, result)
		} else {
			workerResults := make([]WorkerResult, 1)
			workerResults[0] = result
			overall[result.WorkerID] = workerResults
		}
	}
	whole <- overall
}

func get(url string, okState, workerID int) WorkerResult {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create request for %s: %v", url, err)
		return WorkerResult{false, 0.0, workerID}
	}
	start := time.Now()
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "perform request for %s: %v", url, err)
		return WorkerResult{false, time.Since(start), workerID}
	}
	defer res.Body.Close()
	return WorkerResult{res.StatusCode == okState, time.Since(start), workerID}
}

func stats(results Results) Stats {
	var total time.Duration
	var stats Stats
	var durations []time.Duration
	var ok, nok int
	for _, rs := range results {
		for _, r := range rs {
			if r.OK {
				durations = append(durations, r.Time)
				total += r.Time
				ok++
			} else {
				nok++
			}
		}
	}
	stats.Total = ok + nok
	stats.Passed = ok
	stats.Failed = nok
	if ok > 0 {
		stats.Mean = total / time.Duration(ok)
		sort.Slice(durations, func(l, r int) bool { return l < r })
		n := len(durations) / 2
		stats.Perc25 = durations[n/4]
		stats.Perc50 = durations[n/2]
		stats.Perc75 = durations[n/4*3]
	}
	return stats
}
