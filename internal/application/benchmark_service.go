package application

import (
	"context"
	"errors"
	"math"
	"sort"
	"sync"
	"time"

	"api.request.app.backend/internal/domain"
	"github.com/google/uuid"
)

type BenchmarkInput struct {
	FlowID        uuid.UUID
	EnvironmentID *uuid.UUID
	Iterations    int
	Concurrency   int
	DelayMS       int
}

type BenchmarkResult struct {
	TotalRuns     int            `json:"total_runs"`
	PassedRuns    int            `json:"passed_runs"`
	FailedRuns    int            `json:"failed_runs"`
	TotalRequests int            `json:"total_requests"`
	DurationMS    int64          `json:"duration_ms"`
	RPS           float64        `json:"rps"`
	MinLatencyMS  int64          `json:"min_latency_ms"`
	MaxLatencyMS  int64          `json:"max_latency_ms"`
	AvgLatencyMS  float64        `json:"avg_latency_ms"`
	P50LatencyMS  int64          `json:"p50_latency_ms"`
	P90LatencyMS  int64          `json:"p90_latency_ms"`
	P95LatencyMS  int64          `json:"p95_latency_ms"`
	Errors        map[string]int `json:"errors,omitempty"`
}

type BenchmarkService interface {
	RunBenchmark(ctx context.Context, ownerID uuid.UUID, input BenchmarkInput) (*BenchmarkResult, error)
}

type benchmarkService struct {
	flowRepo       domain.FlowRepository
	workspaceRepo  domain.WorkspaceRepository
	requestService RequestService
}

func NewBenchmarkService(
	flowRepo domain.FlowRepository,
	workspaceRepo domain.WorkspaceRepository,
	requestService RequestService,
) BenchmarkService {
	return &benchmarkService{
		flowRepo:       flowRepo,
		workspaceRepo:  workspaceRepo,
		requestService: requestService,
	}
}

func (s *benchmarkService) RunBenchmark(ctx context.Context, ownerID uuid.UUID, input BenchmarkInput) (*BenchmarkResult, error) {
	flow, err := s.flowRepo.GetByID(ctx, input.FlowID)
	if err != nil {
		return nil, errors.New("flow not found")
	}
	if _, err := s.workspaceRepo.GetByID(ctx, flow.WorkspaceID, ownerID); err != nil {
		return nil, err
	}

	requests, err := s.requestService.GetRequestsByFlow(ctx, ownerID, input.FlowID)
	if err != nil {
		return nil, err
	}
	if len(requests) == 0 {
		return nil, errors.New("cannot benchmark flow with no requests")
	}

	iterations := input.Iterations
	if iterations <= 0 {
		iterations = 5
	}
	if iterations > 100 {
		iterations = 100
	}

	concurrency := input.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}
	if concurrency > 10 {
		concurrency = 10
	}

	type runOutcome struct {
		passed    bool
		duration  int64
		errString string
	}

	jobs := make(chan int, iterations)
	results := make(chan runOutcome, iterations)

	startTime := time.Now()
	var wg sync.WaitGroup

	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				iterStart := time.Now()
				iterPassed := true
				var iterErr string
				accumulatedVars := make(map[string]string)

				for _, req := range requests {
					res, runErr := s.requestService.ExecuteRequestWithVariables(ctx, ownerID, req.ID, input.EnvironmentID, accumulatedVars)
					if runErr != nil {
						iterPassed = false
						iterErr = runErr.Error()
						break
					}
					if res.Status >= 400 || res.TestsFailed > 0 {
						iterPassed = false
						if res.TestsFailed > 0 {
							iterErr = "assertion failed"
						} else {
							iterErr = res.StatusText
						}
						break
					}
					for k, v := range res.ExtractedVariables {
						accumulatedVars[k] = v
					}
				}

				results <- runOutcome{
					passed:    iterPassed,
					duration:  time.Since(iterStart).Milliseconds(),
					errString: iterErr,
				}

				if input.DelayMS > 0 {
					time.Sleep(time.Duration(input.DelayMS) * time.Millisecond)
				}
			}
		}()
	}

	for i := 0; i < iterations; i++ {
		jobs <- i
	}
	close(jobs)

	wg.Wait()
	close(results)

	totalDuration := time.Since(startTime).Milliseconds()
	passedRuns := 0
	failedRuns := 0
	var latencies []int64
	var sumLatency int64
	errorCounts := make(map[string]int)

	for r := range results {
		if r.passed {
			passedRuns++
		} else {
			failedRuns++
			if r.errString != "" {
				errorCounts[r.errString]++
			}
		}
		latencies = append(latencies, r.duration)
		sumLatency += r.duration
	}

	sort.Slice(latencies, func(i, j int) bool {
		return latencies[i] < latencies[j]
	})

	var minLatency, maxLatency, p50, p90, p95 int64
	var avgLatency float64
	if len(latencies) > 0 {
		minLatency = latencies[0]
		maxLatency = latencies[len(latencies)-1]
		avgLatency = float64(sumLatency) / float64(len(latencies))
		p50 = latencies[percentileIndex(len(latencies), 0.50)]
		p90 = latencies[percentileIndex(len(latencies), 0.90)]
		p95 = latencies[percentileIndex(len(latencies), 0.95)]
	}

	totalReqs := iterations * len(requests)
	var rps float64
	if totalDuration > 0 {
		rps = math.Round((float64(totalReqs)/(float64(totalDuration)/1000.0))*100) / 100
	}

	return &BenchmarkResult{
		TotalRuns:     iterations,
		PassedRuns:    passedRuns,
		FailedRuns:    failedRuns,
		TotalRequests: totalReqs,
		DurationMS:    totalDuration,
		RPS:           rps,
		MinLatencyMS:  minLatency,
		MaxLatencyMS:  maxLatency,
		AvgLatencyMS:  math.Round(avgLatency*100) / 100,
		P50LatencyMS:  p50,
		P90LatencyMS:  p90,
		P95LatencyMS:  p95,
		Errors:        errorCounts,
	}, nil
}

func percentileIndex(count int, p float64) int {
	idx := int(math.Ceil(float64(count)*p)) - 1
	if idx < 0 {
		return 0
	}
	if idx >= count {
		return count - 1
	}
	return idx
}
