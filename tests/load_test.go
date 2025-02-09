package tests

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"testing"
	"time"
)

var (
	apiURL   = "http://localhost:1123/metrics"
	types    = []string{"free", "standard", "advanced", "pro"}
	versions = []string{"v1", "v2", "v3"}
	rateLimit = 30 // QPS
)

type JobRequest struct {
	Version string `json:"version"`
	Type    string `json:"type"`
}

func sendRequest(client *http.Client, wg *sync.WaitGroup) {
	defer wg.Done()

	job := JobRequest{
		Version: versions[rand.Intn(len(versions))],
		Type:    types[rand.Intn(len(types))],
	}

	data, _ := json.Marshal(job)
	resp, err := client.Post(apiURL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		log.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Non-200 response:", resp.Status)
	} else {
		log.Println("Request sent successfully:", job)
	}
}

func TestLoadTest(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	client := &http.Client{Timeout: 5 * time.Second}
	wg := &sync.WaitGroup{}
	ticker := time.NewTicker(time.Second / time.Duration(rateLimit))
	defer ticker.Stop()

	log.Printf("Starting load test with %d QPS\n", rateLimit)

	for i := 0; i < 100; i++ { // Run for 100 iterations
		wg.Add(1)
		go sendRequest(client, wg)
		<-ticker.C // Wait for the next tick
	}

	wg.Wait()
}
