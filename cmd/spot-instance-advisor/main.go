package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

type SpotPrice struct {
	InstanceType string `json:"InstanceType"`
	SpotPrice    string `json:"SpotPrice"`
}

type InterruptData struct {
	Savings  int `json:"s"` // Savings
	IntrFreq int `json:"r"` // Interrupt Frequency
}

type LabelInfo struct {
	Index int    `json:"index"`
	Label string `json:"label"`
	Dots  int    `json:"dots"`
	Max   int    `json:"max"`
}

// root of JSON
// {"spot_advisor": {"us-west-2": {"Linux": {}}} }
type AdvisorData struct {
	Ranges      []LabelInfo                                    `json:"ranges"`
	SpotAdvisor map[string]map[string]map[string]InterruptData `json:"spot_advisor"`
	// ignore other keys
}

func runCommandWithArgs(command []string) (string, error) {
	cmd := exec.Command(command[0], command[1:]...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("Command execution failed: %s", err)
	}
	return out.String(), nil
}

func removeDuplicates(price []SpotPrice) []SpotPrice {
	seen := make(map[string]bool)
	result := []SpotPrice{}

	for _, price := range price {
		if !seen[price.InstanceType] {
			seen[price.InstanceType] = true
			result = append(result, price)
		}
	}

	return result
}

func getInstanceTypes(region string, minVcpus, maxVcpus, minMemory, maxMemory int) ([]string, error) {
	var vcpuValues []string
	var memoryValues []string
	for i := minVcpus; i <= maxVcpus; i++ {
		vcpuValues = append(vcpuValues, fmt.Sprintf("%d", i))
	}
	for i := minMemory; i <= maxMemory; i++ {
		memoryValues = append(memoryValues, fmt.Sprintf("%d", i*1024))
	}
	vcpuSpec := strings.Join(vcpuValues, ",")
	memorySpec := strings.Join(memoryValues, ",")
	vcpuFilter := fmt.Sprintf("Name=vcpu-info.default-vcpus,Values=%s", vcpuSpec)
	memoryFilter := fmt.Sprintf("Name=memory-info.size-in-mib,Values=%s", memorySpec)

	// get instance types with 4-8 vCPUs and 16-32 GB memory
	command := []string{"aws", "ec2", "describe-instance-types",
		"--filters", vcpuFilter, memoryFilter,
		"--region", region,
		"--query", "InstanceTypes[*].InstanceType", "--output", "json"}
	output, err := runCommandWithArgs(command)
	if err != nil {
		return nil, fmt.Errorf("Error running command: %s", err)
	}

	var instanceTypes []string
	err = json.Unmarshal([]byte(output), &instanceTypes)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling JSON: %s", err)
	}

	return instanceTypes, nil
}

func getSpotPrices(instanceTypes []string, region string) ([]SpotPrice, error) {
	command_head := []string{"aws", "ec2", "describe-spot-price-history",
		"--product-descriptions", "Linux/UNIX",
		"--start-time", time.Now().UTC().Format(time.RFC3339),
		"--query", "SpotPriceHistory[*].{InstanceType:InstanceType,SpotPrice:SpotPrice}",
		"--output", "json",
		"--region", region,
		"--instance-types"}
	output, err := runCommandWithArgs(append(command_head, instanceTypes...))
	if err != nil {
		return nil, fmt.Errorf("Error running command: %s", err)
	}
	var prices []SpotPrice
	err = json.Unmarshal([]byte(output), &prices)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling JSON: %s", err)
	}

	// remove duplicates
	uniqed := removeDuplicates(prices)

	return uniqed, nil
}

// "ranges" is a list of objects as below
//
//	[{"index": 0, "label": "<5%", ...}, {"index": 1, "label": "5-10%", ...}, ...]
//
// then, create a map that key is "index" and value is "label"
//
//	{0: "<5%", 1: "5-10%", ...}
func createRangesMap(ranges []LabelInfo) map[int]string {
	rangesMap := make(map[int]string)
	for _, item := range ranges {
		rangesMap[item.Index] = item.Label
	}
	return rangesMap
}

func getIntrrupData(region string) (map[string]InterruptData, map[int]string, error) {
	type InterruptDataString struct {
	}

	// fetch spot advisor data
	// this URL is from https://aws.amazon.com/ec2/spot/instance-advisor/
	advisorDataUrl := "https://spot-bid-advisor.s3.amazonaws.com/spot-advisor-data.json"
	resp, err := http.Get(advisorDataUrl)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, err
	}

	var advisorData AdvisorData
	err = json.NewDecoder(resp.Body).Decode(&advisorData)
	if err != nil {
		return nil, nil, err
	}

	// same as `jq -r '.spot_advisor.["us-west-2"].Linux'``
	interruptData := advisorData.SpotAdvisor[region]["Linux"]

	return interruptData, createRangesMap(advisorData.Ranges), nil
}

func main() {
	region := flag.String("region", "us-west-2", "AWS region")
	minVcpus := flag.Int("min-vcpus", 4, "Number of CPUs")
	maxVcpus := flag.Int("max-vcpus", 4, "Number of CPUs")
	minMemory := flag.Int("min-memory", 16, "Memory in GB")
	maxMemory := flag.Int("max-memory", 32, "Memory in GB")

	flag.Parse()

	instanceTypes, err := getInstanceTypes(*region, *minVcpus, *maxVcpus, *minMemory, *maxMemory)
	if err != nil {
		log.Fatalf("Error getting instance types: %s", err)
	}

	prices, err := getSpotPrices(instanceTypes, *region)
	if err != nil {
		log.Fatalf("Error getting spot prices: %s", err)
	}

	interruptData, rangesMap, err := getIntrrupData(*region)
	if err != nil {
		log.Fatalf("Error getting interrupt data: %s", err)
	}

	// join instanceTypes with prices and output it
	fmt.Printf("%-20s %-10s %-10s %-10s\n", "InstanceType", "SpotPrice", "Savings", "IntrFreq")
	for _, price := range prices {
		if data, ok := interruptData[price.InstanceType]; ok {
			fmt.Printf("%-20s %-10s %-10d %-10s\n", price.InstanceType, price.SpotPrice, data.Savings, rangesMap[data.IntrFreq])
		}
	}
}
