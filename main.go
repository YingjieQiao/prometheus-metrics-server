package main

import (
	"flag"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	types   = []string{"free", "standard", "advanced", "pro"}
	versions = []string{"v1", "v2", "v3"}
	workers = 0

	// curl "http://localhost:9090/api/v1/query?query=worker_jobs_processed_total"
	processedCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "worker",
			Subsystem: "jobs",
			Name:      "processed_total",
			Help:      "Total number of jobs processed by the workers",
		},
		[]string{"worker_id", "type", "version"},
	)

	pendingCounterVec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "worker",
			Subsystem: "jobs",
			Name:      "pending",
			Help:      "Number of pending jobs",
		},
		[]string{"type", "version"},
	)

	processingTimeVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "worker",
			Subsystem: "jobs",
			Name:      "process_time_seconds",
			Help:      "Amount of time spent processing jobs",
		},
		[]string{"worker_id", "type", "version"},
	)
)

func init() {
	flag.IntVar(&workers, "workers", 10, "Number of workers to use")
}

func getType() string {
	return types[rand.Int()%len(types)]
}

func getVersion() string {
	return versions[rand.Int()%len(versions)]
}

func main() {
	flag.Parse()

	// register with the prometheus collector
	prometheus.MustRegister(
		processedCounterVec,
		pendingCounterVec,
		processingTimeVec,
	)

	// create a channel with a 10,000 Job buffer
	jobsChannel := make(chan *Job, 10000)

	// start the job processor
	go startJobProcessor(jobsChannel)

	go createJobs(jobsChannel)

	handler := http.NewServeMux()
	handler.Handle("/metrics", prometheus.Handler())

	log.Println("[INFO] starting HTTP server on port :9009")
	log.Fatal(http.ListenAndServe(":1123", handler))
}

type Job struct {
	Type  string
	Version string
	Sleep time.Duration
}

// makeJob creates a new job with a random sleep time between 10 ms and 4000ms
func makeJob() *Job {
	return &Job{
		Type:  getType(),
		Version: getVersion(),
		Sleep: time.Duration(rand.Int()%100+10) * time.Millisecond,
	}
}

func startJobProcessor(jobs <-chan *Job) {
	log.Printf("[INFO] starting %d workers\n", workers)
	wait := sync.WaitGroup{}
	// notify the sync group we need to wait for 10 goroutines
	wait.Add(workers)

	// start 10 works
	for i := 0; i < workers; i++ {
		go func(workerID int) {
			// start the worker
			startWorker(workerID, jobs)
			wait.Done()
		}(i)
	}

	wait.Wait()
}

func createJobs(jobs chan<- *Job) {
	for {
		// create a random job
		job := makeJob()
		// track the job in the pending tracker
		pendingCounterVec.WithLabelValues(job.Type, job.Version).Inc()
		// send the job down the channel
		jobs <- job
		// don't pile up too quickly
		time.Sleep(5 * time.Millisecond)
	}
}

// creates a worker that pulls jobs from the job channel
func startWorker(workerID int, jobs <-chan *Job) {
	for {
		select {
		// read from the job channel
		case job := <-jobs:
			startTime := time.Now()

			// mock processing the request
			time.Sleep(job.Sleep)
			log.Printf("[%d][%s] Processed job in %0.3f seconds", workerID, job.Type, time.Now().Sub(startTime).Seconds())
			// track the total number of jobs processed by the worker
			processedCounterVec.WithLabelValues(strconv.FormatInt(int64(workerID), 10), job.Type, job.Version).Inc()
			// decrement the pending tracker
			pendingCounterVec.WithLabelValues(job.Type, job.Version).Dec()

			processingTimeVec.WithLabelValues(strconv.FormatInt(int64(workerID), 10), job.Type, job.Version).Observe(time.Now().Sub(startTime).Seconds())
		}
	}
}
