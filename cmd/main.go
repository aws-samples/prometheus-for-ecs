package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws-samples/prometheus-for-ecs/pkg/aws"
)

const (
	TARGETS = "/prometheus-targets"
	PORT    = 9001
)

var present bool
var configFileDir, configReloadFrequency string
var prometheusConfigFilePath, scrapeConfigFilePath string

func main() {
	aws.InitializeAWSSession()
	sdMode, present := os.LookupEnv("SERVICE_DISCOVERY_MODE")
	if !present {
		sdMode = "FILE_BASED"
	}
	if sdMode == "FILE_BASED" {
		fileBasedSD()
	} else if sdMode == "HTTP_BASED" {
		httpBasedSD()
	} else {
		log.Printf("Invalid service discovery mode %s", sdMode)
	}
}

func httpBasedSD() {
	log.Println("Service discovery application started in HTTP-based mode")
	serveMux := http.NewServeMux()
	serveMux.HandleFunc(TARGETS, getScrapeConfig)

	stopChannel := make(chan string)
	defer close(stopChannel)

	go func(doneChannel chan string) {
		port := PORT
		addr := fmt.Sprintf(":%d", port)
		fmt.Printf("Started HTTP server at %s\n", addr)

		server := &http.Server{
			Addr:           addr,
			Handler:        serveMux,
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		server.SetKeepAlivesEnabled(true)
		log.Fatal(server.ListenAndServe())
		doneChannel <- "HTTP server terminated abnormally"
	}(stopChannel)

	fmt.Println("Waiting for all goroutines to complete")

	for {
		select {
		case status := <-stopChannel:
			fmt.Println(status)
			break
		}
	}
}

func fileBasedSD() {
	log.Println("Service discovery application started in file-based mode")
	aws.InitializeAWSSession()

	configFileDir, present = os.LookupEnv("CONFIG_FILE_DIR")
	if !present {
		configFileDir = "/etc/config/"
	}
	configReloadFrequency, present = os.LookupEnv("CONFIG_RELOAD_FREQUENCY")
	if !present {
		configReloadFrequency = "30"
	}

	loadPrometheusConfig()
	initScrapeTargetConfig()

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
	log.Println("Periodic reloads under progress...")

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
	// When deployed with OTel Collector, Prometheus runs as a Receiver in the OTel Pipeline
	// Hence, Prometheus configuration is part of the OTel Pipeline configuration which is loaded from SSM
	deploymentModeParameter, present := os.LookupEnv("DEPLOYMENT_MODE")
	if present && (deploymentModeParameter == "OTEL" || deploymentModeParameter == "otel") {
		log.Println("Running in OTel mode")
		return
	}

	prometheusConfigParameter, present := os.LookupEnv("PROMETHEUS_CONFIG_PARAMETER_NAME")
	if !present {
		prometheusConfigParameter = "ECS-Prometheus-Configuration"
	}
	prometheusConfig := aws.GetParameter(prometheusConfigParameter)

	prometheusConfigFilePath = strings.Join([]string{configFileDir, "prometheus.yaml"}, "/")
	err := ioutil.WriteFile(prometheusConfigFilePath, []byte(*prometheusConfig), 0644)
	if err != nil {
		log.Println(err)
	}
	log.Println("Loaded Prometheus configuration file")

}

func initScrapeTargetConfig() {
	scrapeConfigFilePath = strings.Join([]string{configFileDir, "ecs-services.json"}, "/")
	err := ioutil.WriteFile(scrapeConfigFilePath, []byte("[]"), 0644)
	if err != nil {
		log.Println(err)
	}
	log.Println("Created initial scrape target configuration file")
}

func reloadScrapeConfig() {
	scrapConfig := buildSrapeConfig()
	err := ioutil.WriteFile(scrapeConfigFilePath, []byte(*scrapConfig), 0644)
	if err != nil {
		log.Println(err)
	}
}

func getScrapeConfig(w http.ResponseWriter, r *http.Request) {
	scrapConfig := buildSrapeConfig()
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, *scrapConfig)
}

func buildSrapeConfig() *string {
	discoveryNamespacesParameter, present := os.LookupEnv("DISCOVERY_NAMESPACES_PARAMETER_NAME")
	if !present {
		discoveryNamespacesParameter = "ECS-ServiceDiscovery-Namespaces"
	}
	namespaceList := aws.GetParameter(discoveryNamespacesParameter)
	namespaces := strings.Split(*namespaceList, ",")
	scrapConfig := aws.GetPrometheusScrapeConfig(namespaces)
	return scrapConfig
}
