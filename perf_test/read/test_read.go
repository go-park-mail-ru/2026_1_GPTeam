package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	vegeta "github.com/tsenart/vegeta/v12/lib"
)

func main() {
	err := godotenv.Load("perf_test/.env")
	if err != nil {
		fmt.Println("Error loading .env file: ", err)
		return
	}

	host := os.Getenv("HOST")
	targetURL := host + "/api/transactions"
	totalRequestsStr := os.Getenv("TOTAL_REQUESTS")
	totalRequests, err := strconv.Atoi(totalRequestsStr)
	if err != nil {
		fmt.Println("Error converting TOTAL_REQUESTS to int: ", err)
		return
	}
	requestsPerSecondStr := os.Getenv("REQUEST_PER_SECOND")
	requestsPerSecond, err := strconv.Atoi(requestsPerSecondStr)
	if err != nil {
		fmt.Println("Error converting REQUEST_PER_SECOND to int: ", err)
		return
	}
	attackDuration := time.Duration(totalRequests/requestsPerSecond) * time.Second
	targets := make([]vegeta.Target, totalRequests)
	token := os.Getenv("TOKEN")
	csrfToken := os.Getenv("CSRF_TOKEN")
	cookies := []http.Cookie{
		{
			Name:  "token",
			Value: token,
			Path:  "/",
		},
		{
			Name:  "csrf_token",
			Value: csrfToken,
			Path:  "/",
		},
	}
	for i := 0; i < totalRequests; i++ {
		headers := map[string][]string{
			"Content-Type": {"application/json"},
			"X-CSRF-Token": {csrfToken},
		}
		var cookieString string
		for _, c := range cookies {
			if cookieString != "" {
				cookieString += "; "
			}
			cookieString += fmt.Sprintf("%s=%s", c.Name, c.Value)
		}
		headers["Cookie"] = []string{cookieString}
		targets[i] = vegeta.Target{
			Method: http.MethodGet,
			URL:    targetURL + "/" + strconv.Itoa(i+1),
			Body:   nil,
			Header: headers,
		}
	}

	rate := vegeta.Rate{Freq: requestsPerSecond, Per: time.Second}
	targeter := vegeta.NewStaticTargeter(targets...)
	attacker := vegeta.NewAttacker()

	fmt.Printf("Starting attack: %d requests at %d rps\n", totalRequests, requestsPerSecond)
	startTime := time.Now()
	var metrics vegeta.Metrics
	for res := range attacker.Attack(targeter, rate, attackDuration, "Read 100k transactions") {
		metrics.Add(res)
	}
	metrics.Close()
	duration := time.Since(startTime)

	fmt.Println("\nResults:")
	fmt.Printf("  Success rate:  %.2f%%\n", metrics.Success*100)
	fmt.Printf("  Total requests: %d\n", metrics.Requests)
	fmt.Printf("  Total duration:      %s\n", duration)
	fmt.Printf("  Actual rate:   %.2f req/s\n", metrics.Rate)
	fmt.Printf("  Throughput:    %.2f req/s\n", metrics.Throughput)
	fmt.Printf("Latencies:\n")
	fmt.Printf("  Min:  %s\n", metrics.Latencies.Min)
	fmt.Printf("  Mean: %s\n", metrics.Latencies.Mean)
	fmt.Printf("  50th: %s\n", metrics.Latencies.P50)
	fmt.Printf("  90th: %s\n", metrics.Latencies.P90)
	fmt.Printf("  Max:  %s\n", metrics.Latencies.Max)
	fmt.Printf("  Bytes (In/Out): %.0f / %.0f\n", metrics.BytesIn.Mean, metrics.BytesOut.Mean)
	fmt.Printf("  Status codes: %v\n", metrics.StatusCodes)
}
