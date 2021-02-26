package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws-samples/prometheus-for-ecs/aws"
)

const (
	PrometheusConfigParameter    = "ECS-Prometheus-Configuration"
	DiscoveryNamespacesParameter = "ECS-ServiceDiscovery-Namespaces"
)

var prometheusConfigFilePath string
var scrapeConfigFilePath string

func main() {
	aws.InitializeAWSSession()

	configFileDir, present := os.LookupEnv("CONFIG_FILE_DIR")
	if !present {
		configFileDir = "/etc/config/"
	}
	configReloadFrequency, present := os.LookupEnv("CONFIG_RELOAD_FREQUENCY")
	if !present {
		configReloadFrequency = "30"
	}
	prometheusConfigFilePath = strings.Join([]string{configFileDir, "prometheus.yaml"}, "/")
	scrapeConfigFilePath = strings.Join([]string{configFileDir, "ecs-services.json"}, "/")

	loadPrometheusConfig()
	loadScrapeConfig()

	go func() {
		reloadFrequency, _ := strconv.Atoi(configReloadFrequency)
		ticker := time.NewTicker(time.Duration(reloadFrequency) * time.Second)
		for {
			select {
			case <-ticker.C:
				//
				// Ticker contains a channel
				// It sends the time on the channel after the number of ticks specified by the duration have elapsed.
				//
				reloadScrapeConfig()
			}
		}
	}()

	//
	// Block indefinitely on the main channel
	//
	stopChannel := make(chan string)
	for {
		select {
		case status := <-stopChannel:
			fmt.Println(status)
			break
		}
	}

}

func loadPrometheusConfig() {
	prometheusConfig := aws.GetParameter(PrometheusConfigParameter)
	err := ioutil.WriteFile(prometheusConfigFilePath, []byte(*prometheusConfig), 0644)
	if err != nil {
		log.Println(err)
	}
}

func loadScrapeConfig() {
	err := ioutil.WriteFile(scrapeConfigFilePath, []byte("[]"), 0644)
	if err != nil {
		log.Println(err)
	}
}

func reloadScrapeConfig() {
	namespaceList := aws.GetParameter(DiscoveryNamespacesParameter)
	namespaces := strings.Split(*namespaceList, ",")
	scrapConfig := aws.GetPrometheusScrapeConfig(namespaces)
	err := ioutil.WriteFile(scrapeConfigFilePath, []byte(*scrapConfig), 0644)
	if err != nil {
		log.Println(err)
	}
}
