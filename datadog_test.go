package main

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func assertErrorMsg(t *testing.T, errMsg string) {
	if err := recover(); err != nil {
		if e, ok := err.(error); ok {
			assert.EqualError(t, e, errMsg)
		} else {
			assert.Equal(t, errMsg, err)
		}
	} else {
		t.Fatal("validateType didn't panic as expected")
	}
}

func TestExpandPath(t *testing.T) {
	assert.Equal(t, "foo.txt", expandPath("foo.txt"))
	assert.Equal(t, os.Getenv("HOME")+"/foo.txt", expandPath("~/foo.txt"))
	assert.Equal(t, "foo/~/bar.txt", expandPath("foo/~/bar.txt"))
}

func TestReadDatadogKeysWithEnvVars(t *testing.T) {
	apiKey, appKey := "myApiKey", "myAppKey"
	os.Setenv("DATADOG_API_KEY", apiKey)
	os.Setenv("DATADOG_APP_KEY", appKey)
	keysChan := make(chan string, 2)
	go readDatadogKeys("/tmp/do-not-exist", keysChan)
	assert.Equal(t, apiKey, <-keysChan)
	assert.Equal(t, appKey, <-keysChan)
}

func TestReadDatadogKeysWithConfigFile(t *testing.T) {
	os.Setenv("DATADOG_API_KEY", "")
	os.Setenv("DATADOG_APP_KEY", "")
	keysChan := make(chan string, 2)
	go readDatadogKeys("datadogrc.example.json", keysChan)
	assert.Equal(t, "myApiKeyInJson", <-keysChan)
	assert.Equal(t, "myAppKeyInJson", <-keysChan)
}

func TestValidateType(t *testing.T) {
	assert.Equal(t, "counter", validateType("increment"))
	assert.Equal(t, "counter", validateType("incr"))
	assert.Equal(t, "counter", validateType("i"))
	assert.Equal(t, "counter", validateType("counter"))
	assert.Equal(t, "counter", validateType("c"))

	assert.Equal(t, "gauge", validateType("gauge"))
	assert.Equal(t, "gauge", validateType("g"))
}

func TestValidateTypePanics(t *testing.T) {
	defer assertErrorMsg(t, `'histogram' is not a valid metric type. must be one of 'counter', or 'gauge'.`)
	validateType("histogram")
}

func TestCreateDataPoint(t *testing.T) {
	point := createDataPoint("100")
	assert.Equal(t, float64(100), point[1])
	assert.Equal(t, float64(time.Now().Unix()), point[0])
}

func TestCreateDataPointPanics(t *testing.T) {
	defer assertErrorMsg(t, `strconv.ParseFloat: parsing "i'm no float": invalid syntax`)
	createDataPoint("i'm no float")
}

func TestValidateAndConvertPoints(t *testing.T) {
	points := validateAndConvertPoints([]string{"12.45"})
	assert.Equal(t, float64(time.Now().Unix()), points[0][0])
	assert.Equal(t, float64(12.45), points[0][1])

	points = validateAndConvertPoints([]string{"125.20", "329.89683"})
	assert.Equal(t, float64(time.Now().Unix()), points[0][0])
	assert.Equal(t, float64(125.2), points[0][1])
	assert.Equal(t, float64(time.Now().Unix()), points[1][0])
	assert.Equal(t, float64(329.89683), points[1][1])
}

func TestValidateAndConvertPointsPanics(t *testing.T) {
	defer assertErrorMsg(t, "no value(s)")
	validateAndConvertPoints([]string{})
}

func TestParseArgsPanicsWithNoArgs(t *testing.T) {
	defer assertErrorMsg(t, "not enough arguments. usage: datadog TYPE METRIC VALUE(S)...")
	parseArgs([]string{}, "")
}

func TestParseArgsPanicsWithOneArg(t *testing.T) {
	defer assertErrorMsg(t, "not enough arguments. usage: datadog TYPE METRIC VALUE(S)...")
	parseArgs([]string{"gauge"}, "")
}

func TestParseArgsPanicsWithTwoArgs(t *testing.T) {
	defer assertErrorMsg(t, "not enough arguments. usage: datadog TYPE METRIC VALUE(S)...")
	parseArgs([]string{"gauge", "vsco.my_metric"}, "")
}

func TestParseArgsOneValue(t *testing.T) {
	metric := parseArgs([]string{"gauge", "vsco.my_metric", "58.274"}, "project:glory")
	assert.NotNil(t, metric)
	assert.Equal(t, "gauge", metric.Type)
	assert.Equal(t, "vsco.my_metric", metric.Metric)
	assert.Equal(t, []string{"project:glory"}, metric.Tags)
	assert.Equal(t, float64(time.Now().Unix()), metric.Points[0][0])
	assert.Equal(t, float64(58.274), metric.Points[0][1])
}

func TestParseArgsThreeValues(t *testing.T) {
	metric := parseArgs([]string{"gauge", "vsco.my_metric", "58.274", "526.52", "-0.242"}, "project:glory")
	assert.NotNil(t, metric)
	assert.Equal(t, "gauge", metric.Type)
	assert.Equal(t, "vsco.my_metric", metric.Metric)
	assert.Equal(t, []string{"project:glory"}, metric.Tags)
	assert.Equal(t, float64(time.Now().Unix()), metric.Points[0][0])
	assert.Equal(t, float64(58.274), metric.Points[0][1])
	assert.Equal(t, float64(time.Now().Unix()), metric.Points[1][0])
	assert.Equal(t, float64(526.52), metric.Points[1][1])
	assert.Equal(t, float64(time.Now().Unix()), metric.Points[2][0])
	assert.Equal(t, float64(-0.242), metric.Points[2][1])
}
