package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"strconv"
	"strings"
	"time"

	"github.com/zorkian/go-datadog-api"
)

var homeTildeShortcut = "~" + string(os.PathSeparator)

type datadogKeys struct {
	ApiKey string `json:"api_key"`
	AppKey string `json:"app_key"`
}

func init() {
	log.SetOutput(os.Stderr)
}

func expandPath(path string) string {
	if path[:2] == homeTildeShortcut {
		currentUser, err := user.Current()
		if err != nil {
			panic(err)
		}
		return strings.Replace(path, homeTildeShortcut, currentUser.HomeDir+string(os.PathSeparator), 1)
	} else {
		return path
	}
}

func readDatadogKeys(configFilePath string, keysChan chan<- string) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("error: datadog configuration missing or inaccessible.")
			log.Fatalf("error: %v", err)
		}
	}()

	keys := &datadogKeys{
		ApiKey: os.Getenv("DATADOG_API_KEY"),
		AppKey: os.Getenv("DATADOG_APP_KEY"),
	}

	if keys.ApiKey == "" || keys.AppKey == "" {
		f, err := os.Open(expandPath(configFilePath))
		if err != nil {
			panic(err)
		}
		defer f.Close()

		err = json.NewDecoder(f).Decode(keys)
		if err != nil {
			panic(err)
		}
	}

	keysChan <- keys.ApiKey
	keysChan <- keys.AppKey
}

func validateType(metricType string) string {
	switch metricType {
	case "increment", "incr", "i", "c", "counter":
		return "counter"
	case "gauge", "g":
		return "gauge"
	default:
		panic(fmt.Sprintf("'%s' is not a valid metric type. must be one of 'counter', or 'gauge'.", metricType))
	}
}

func createDataPoint(value string) datadog.DataPoint {
	converted, err := strconv.ParseFloat(value, 64)
	if err != nil {
		panic(err)
	}
	now := float64(time.Now().Unix())
	return datadog.DataPoint([2]float64{now, converted})
}

func validateAndConvertPoints(data []string) (points []datadog.DataPoint) {
	switch len(data) {
	case 0:
		panic("no value(s)")
	default:
		for _, v := range data {
			points = append(points, createDataPoint(v))
		}
	}
	return points
}

func parseArgs(args []string, tags string) datadog.Metric {
	if len(args) < 3 {
		panic("not enough arguments. usage: datadog TYPE METRIC VALUE(S)...")
	}

	return datadog.Metric{
		Type:   validateType(args[0]),
		Metric: args[1],
		Points: validateAndConvertPoints(args[2:]),
		Tags:   strings.Split(tags, ","),
	}
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("error: %v", err)
		}
	}()

	// datadog gauge my_metric values...
	var dryRun bool
	flag.BoolVar(&dryRun, "dry-run", false, "Don't send data to datadog.")
	var tags string
	flag.StringVar(&tags, "tags", "", "Tags to add to this metric, e.g. 'key:value,key2:value2'.")
	var configFile string
	flag.StringVar(&configFile, "conf", "~/.datadogrc", "Datadog app and api keys")
	flag.Parse()

	keysChan := make(chan string, 2)
	go readDatadogKeys(configFile, keysChan)

	metric := parseArgs(flag.Args(), tags)

	if dryRun {
		encoded, err := json.Marshal(metric)
		if err != nil {
			panic(err)
		}
		log.Println(string(encoded))
	} else {
		client := datadog.NewClient(<-keysChan, <-keysChan)
		err := client.PostMetrics([]datadog.Metric{metric})
		if err != nil {
			panic(err)
		}
	}
}
