package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/bgentry/speakeasy"
	"github.com/vharitonsky/iniflags"
	"github.com/zorkian/go-datadog-api"
)

var (
	err          error
	startTime    time.Time
	endTime      time.Time
	inputFormat  = "2006/01/02-15:04"
	outputFormat = "2006/01/02-15:04:05"
	interval     time.Duration
	intMap       = map[string]time.Duration{}
)

// execution starts here...
func main() {
	apiKey := flag.String("api_key", "", "API Key to connect to DataDog")
	appKey := flag.String("app_key", "", "APP Key for this app in DataDog")
	query := flag.String("query", "", "The DataDog query to run (required)")
	rangeStart := flag.String("start", "", "The starting point for the date range to query (format: yyyy/mm/dd-hh:mm) (required)")
	rangeEnd := flag.String("end", "", "The ending point for the date range to query (format: yyyy/mm/dd-hh:mm) (required)")
	intStr := flag.String("interval", "1h", "The perferred data interval. [5m, 10m, 20m, 30m, 1h, 2h, 4h, 8h, 12h, 24h]")
	//csvFilepath := flag.String("csv_file", "", "The filepath of the CSV file to output to")
	version := flag.Bool("v", false, "Version of the binary (optional)")
	iniflags.Parse()

	// version check
	if *version {
		fmt.Println(Version)
		os.Exit(0)
	}

	// required fields
	if *query == "" {
		log.Fatal("'-query' parameter is required")
	}
	if *rangeStart == "" {
		log.Fatal("'-start' parameter is required")
	} else {
		startTime, err = time.Parse(inputFormat, *rangeStart)
		if err != nil {
			log.Println("Unable to parse 'start' date.  Expected format is: yyyy/mm/dd-hh:mm")
			log.Fatal(err)
		}
	}
	if *rangeEnd == "" {
		log.Fatal("'-end' parameter is required")
	} else {
		endTime, err = time.Parse(inputFormat, *rangeEnd)
		if err != nil {
			log.Println("Unable to parse 'end' date.  Expected format is: yyyy/mm/dd-hh:mm")
			log.Fatal(err)
		}
	}

	// initialize the interval mapping to a query duration
	intMap["5m"], _ = time.ParseDuration("24h")    // 1d
	intMap["10m"], _ = time.ParseDuration("48h")   // 2d
	intMap["20m"], _ = time.ParseDuration("96h")   // 4d
	intMap["30m"], _ = time.ParseDuration("144h")  // 6d
	intMap["1h"], _ = time.ParseDuration("288h")   // 12d
	intMap["2h"], _ = time.ParseDuration("576h")   // 24d
	intMap["4h"], _ = time.ParseDuration("1152h")  // 48d
	intMap["8h"], _ = time.ParseDuration("2304h")  // 96d
	intMap["12h"], _ = time.ParseDuration("3456h") // 144d
	intMap["24h"], _ = time.ParseDuration("6912h") // 288d

	// validate that the provided interval value is valid
	if intDur, inIntMap := intMap[*intStr]; inIntMap {
		interval = intDur
	} else {
		log.Fatalf("The '-interval' value '%s' is not one of: 5m, 10m, 20m, 30m, 1h, 2h, 4h, 8h, 12h, 24h", *intStr)
	}

	// figure out the range that will provide an even multiple of the interval
	// important in order to ensure the interval is consistent across the whole dataset
	duration := endTime.Sub(startTime)
	startTime = endTime.Add(-time.Duration(math.Ceil(duration.Hours()/interval.Hours())) * interval)

	// ask for apiKey and app_key if not specified
	if *apiKey == "" {
		apiK, err := speakeasy.Ask("Enter your API_KEY: ")
		if err != nil {
			log.Fatal(err)
		}
		flag.Set("api_key", apiK)
	}
	if *appKey == "" {
		appK, err := speakeasy.Ask("Enter your APP_KEY: ")
		if err != nil {
			log.Fatal(err)
		}
		flag.Set("app_key", appK)
	}

	// logging setup
	_ = os.Remove("datadog_csv.log") // start by deleting the log file...
	logFile, err := os.OpenFile("datadog_csv.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	// lets get working...
	log.Println("Connecting to DataDog...")

	// setup a connect to DataDog
	client := datadog.NewClient(*apiKey, *appKey)

	log.Printf("Requested date range: %s to %s", *rangeStart, *rangeEnd)
	log.Printf("Querying date range: %s to %s", startTime.Format(inputFormat), endTime.Format(inputFormat))

	// get to querying...
	tempTime := startTime
	for tempTime.Before(endTime) {
		log.Printf("Querying '%s' from '%s' to '%s'...\n", *query, tempTime.String(), tempTime.Add(interval))
		details, err := client.QueryMetrics(tempTime.Unix(), tempTime.Add(interval).Unix(), *query)
		if err != nil {
			log.Println("Failed to query DataDog metrics")
			log.Fatal(err)
		}

		for _, data := range details {
			for _, point := range data.Points {
				fmt.Printf("%s, %f\n", time.Unix(int64(*point[0])/1000, 0).Format(outputFormat), *point[1])
			}
		}
		tempTime = tempTime.Add(interval)
	}
}
