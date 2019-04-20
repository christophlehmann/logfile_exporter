package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

var (
	interval   int64
	config     Config
	outputFile string
	metrics    = map[string]Metric{}
)

const metricsPrefix = "logfile_"

func main() {
	configFile := *flag.String("config", "config.yaml", "Configuration file")
	outputFile = *flag.String("output", "output.txt", "Output file")
	interval = *flag.Int64("interval", 5, "Interval (seconds)")
	flag.Parse()

	log.Print("Using configuration ", configFile)
	config = readConfiguration(configFile)

	for _, logfile := range config.Logfiles {
		go watchLogfile(logfile)
	}

	time.Sleep(time.Duration(interval) * time.Second)
	for {
		printMetrics()
		writeMetricsToFile()
		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func watchLogfile(logfile Logfile) {
	log.Print("Start watching ", logfile.Filename)
	for {
		fileHandle, err := os.Open(logfile.Filename)
		checkErr(err)

		fileStats, err := fileHandle.Stat()
		checkErr(err)
		fileSize := fileStats.Size()

		if fileSize < logfile.Position {
			log.Printf("Detected logrotation of %s", logfile.Filename)
			logfile.Position = 0
		} else {
			_, err = fileHandle.Seek(logfile.Position, 0)
			checkErr(err)
		}

		scanner := bufio.NewScanner(fileHandle)
		for scanner.Scan() {
			for _, metric := range logfile.Metrics {
				// Todo: Make this work async. Need fix for concurrent map read/write
				updateMetric(metric, scanner.Text())
			}
		}

		logfile.Position = fileSize

		err = fileHandle.Close()
		checkErr(err)

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func updateMetric(metric Metric, line string) {
	if metric.RegexCompiled.MatchString(line) {

		if xx, ok := metrics[metric.Name]; ok {
			xx.Counter++
			metrics[metric.Name] = xx
		} else {
			metric.Counter++
			metrics[metric.Name] = metric
		}
	}
}

func printMetrics() {
	for _, metric := range metrics {
		log.Printf("%s: %d\n", metric.Name, metric.Counter)
	}
}

func writeMetricsToFile() {
	outputFileHandle, err := os.Create(outputFile)
	checkErr(err)
	for _, metric := range metrics {
		helpLine := fmt.Sprintf("# HELP %s%s %s\n", metricsPrefix, metric.Name, metric.Help)
		typeLine := fmt.Sprintf("# TYPE %s%s untyped\n", metricsPrefix, metric.Name)
		metricLine := fmt.Sprintf("%s%s: %d\n", metricsPrefix, metric.Name, metric.Counter)
		_, err = outputFileHandle.Write([]byte(helpLine + typeLine + metricLine))
	}
	err = outputFileHandle.Close()
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
