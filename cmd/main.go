package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws-samples/prometheus-for-ecs/pkg/aws"
)

const (
	DiscoveryNamespacesParameter = "ECS-ServiceDiscovery-Namespaces"
)

var scrapeConfigFilePath string

func main() {
	log.Println("Prometheus configuration reloader started")
	aws.InitializeAWSSession()

	configFileDir, present := os.LookupEnv("CONFIG_FILE_DIR")
	if !present {
		configFileDir = "/etc/config/"
	}
	configReloadFrequency, present := os.LookupEnv("CONFIG_RELOAD_FREQUENCY")
	if !present {
		configReloadFrequency = "30"
	}
	scrapeConfigFilePath = strings.Join([]string{configFileDir, "ecs-services.json"}, "/")

	initScrapeTargetConfig()
	log.Println("Created initial scrape target configuration file")

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
				updateScrapeTargetConfig()
			}
		}
	}()
	log.Println("Periodic updates of scrape target configuration under progress...")

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

func initScrapeTargetConfig() {
	err := ioutil.WriteFile(scrapeConfigFilePath, []byte("[]"), 0644)
	if err != nil {
		log.Println(err)
	}
}

func updateScrapeTargetConfig() {
	namespaceList := aws.GetParameter(DiscoveryNamespacesParameter)
	namespaces := strings.Split(*namespaceList, ",")
	scrapConfig := aws.GetPrometheusScrapeConfig(namespaces)
	err := ioutil.WriteFile(scrapeConfigFilePath, []byte(*scrapConfig), 0644)
	if err != nil {
		log.Println(err)
	}
}
