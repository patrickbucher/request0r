package main

import (
	"flag"
	"fmt"
	"math"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
)

type WorkerResult struct {
	OK   bool
	Time time.Duration
}

var Percentiles = []int{0, 25, 50, 75, 100}

type Stats struct {
	Total  int
	Passed int
	Failed int
	Mean   time.Duration
	Percs  map[int]time.Duration
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
	fmt.Println("Requests:")
	fmt.Printf("%15s %15s %15s %15s\n", "Total", "Passed", "Failed", "Mean")
	fmt.Printf("%15d %15d %15d %15s\n", s.Total, s.Passed, s.Failed, s.Mean)
	fmt.Println("Percentiles:")
	for _, k := range Percentiles {
		fmt.Printf("%14d%% ", k)
	}
	fmt.Println()
	for _, k := range Percentiles {
		fmt.Printf("%15s ", s.Percs[k])
	}
	fmt.Println()
}

func run(url string, workers, requests, okState int) []WorkerResult {
	collector := make(chan []WorkerResult)
	results := make(chan WorkerResult)

	go collect(collector, results)

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(requests)
		go func(ch chan WorkerResult, id int) {
			for r := 0; r < requests; r++ {
				ch <- get(url, okState, w)
				wg.Done()
			}
		}(results, w)
	}
	wg.Wait()
	close(results)

	return <-collector
}

func collect(whole chan<- []WorkerResult, parts <-chan WorkerResult) {
	overall := make([]WorkerResult, 0)
	for result := range parts {
		overall = append(overall, result)
	}
	whole <- overall
}

func get(url string, okState, workerID int) WorkerResult {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create request for %s: %v", url, err)
		return WorkerResult{false, 0.0}
	}

	start := time.Now()
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "perform request for %s: %v", url, err)
		return WorkerResult{false, time.Since(start)}
	}
	defer res.Body.Close()

	return WorkerResult{res.StatusCode == okState, time.Since(start)}
}

func stats(results []WorkerResult) Stats {
	var total time.Duration
	var stats Stats
	var durations []time.Duration
	var ok int

	for _, r := range results {
		if r.OK {
			durations = append(durations, r.Time)
			total += r.Time
			ok++
		}
	}

	stats.Total = len(results)
	stats.Passed = ok
	stats.Failed = stats.Total - stats.Passed
	if ok > 0 {
		stats.Mean = total / time.Duration(ok)
		sort.Slice(durations, func(l, r int) bool {
			return durations[l] < durations[r]
		})
		n := len(durations) / 2
		stats.Percs = make(map[int]time.Duration)
		for _, p := range Percentiles {
			if p == 0 {
				stats.Percs[p] = durations[0]
			} else {
				ratio := float64(p) / 100.0
				index := int(math.Round(float64(n) * ratio))
				stats.Percs[p] = durations[index]
			}
		}
	}

	return stats
}
