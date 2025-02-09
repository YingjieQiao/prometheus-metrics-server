package main

import (
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// metrics
var (
	types    = []string{"free", "standard", "advanced", "pro"}
	versions = []string{"v1", "v2", "v3"}
	workers  = 10

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

// models
type PostJobMetricsRequest struct {
	Version string `json:"version"`
	Type    string `json:"type"`
}


type Job struct {
	Type    string
	Version string
	Sleep   time.Duration
}

func init() {
	prometheus.MustRegister(
		processedCounterVec,
		pendingCounterVec,
		processingTimeVec,
	)

	go startJobProcessor()
}

var jobsChannel = make(chan *Job, 10000)

func startJobProcessor() {
	log.Printf("[INFO] starting %d workers\n", workers)
	var wait sync.WaitGroup
	wait.Add(workers)

	for i := 0; i < workers; i++ {
		go func(workerID int) {
			startWorker(workerID)
			wait.Done()
		}(i)
	}

	wait.Wait()
}

func startWorker(workerID int) {
	for job := range jobsChannel {
		startTime := time.Now()

		time.Sleep(time.Duration(rand.Int()%100+10) * time.Millisecond)
		log.Printf("[%d][%s] Processed job in %0.3f seconds", workerID, job.Type, time.Now().Sub(startTime).Seconds())

		processedCounterVec.WithLabelValues(strconv.Itoa(workerID), job.Type, job.Version).Inc()

		pendingCounterVec.WithLabelValues(job.Type, job.Version).Dec()

		processingTimeVec.WithLabelValues(strconv.Itoa(workerID), job.Type, job.Version).Observe(time.Now().Sub(startTime).Seconds())
	}
}

func main() {
	r := gin.Default()

	r.POST("/metrics", postJobMetrics)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	log.Println("[INFO] starting HTTP server on port :1123")
	log.Fatal(r.Run(":1123"))
}


func postJobMetrics(c *gin.Context) {
	var request PostJobMetricsRequest

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	job := &Job{
		Type:    request.Type,
		Version: request.Version,
	}

	pendingCounterVec.WithLabelValues(job.Type, job.Version).Inc()

	jobsChannel <- job

	c.JSON(http.StatusOK, gin.H{"status": "Job accepted"})
}

// for testing
func getType() string {
	return types[rand.Int()%len(types)]
}

func getVersion() string {
	return versions[rand.Int()%len(versions)]
}
