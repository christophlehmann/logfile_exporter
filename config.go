package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"regexp"
)

type Logfile struct {
	Filename string `yaml:"filename"`
	Position int64
	Metrics  []Metric `yaml:"metrics"`
}

type Metric struct {
	Name          string `yaml:"name"`
	Help          string `yaml:"help"`
	Regex         string `yaml:"regex"`
	RegexCompiled *regexp.Regexp
	Counter       int64
}

type Config struct {
	Logfiles []Logfile `yaml:"logfiles"`
}

func readConfiguration(configFile string) Config {
	yamlConfigFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal("Could not open config", err)
	}

	var c Config
	err = yaml.Unmarshal(yamlConfigFile, &c)
	if err != nil {
		log.Fatal("Error while parsing yaml config", err)
	}

	precompileRegularExpressions(&c)
	return c
}

func precompileRegularExpressions(c *Config) {
	for logfileIndex, logfile := range c.Logfiles {
		for entryIndex, entry := range logfile.Metrics {
			re := regexp.MustCompile(entry.Regex)
			c.Logfiles[logfileIndex].Metrics[entryIndex].RegexCompiled = re
		}
	}
}
