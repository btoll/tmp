package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func main() {
	instance := os.Getenv("PD_SERVICE")
	if instance == "" {
		exit("No instance name (PD_SERVICE).")
	}

	testDir := os.Getenv("TEST_DIR")
	if testDir == "" {
		exit("No test directory (TEST_DIR).")
	}

	environment := "stg"
	// ex. stg.event-storage-ex-api.service.consul
	instanceName := strings.Join([]string{environment, instance, "service", "consul"}, ".")
	fmt.Printf("[INFO] Instance name: %s\n", instanceName)

	cmd := fmt.Sprintf("dig +short %s SRV | cut -d' ' -f3", instanceName)
	digCmd := exec.Command("sh", "-c", cmd)
	port, err := digCmd.CombinedOutput()
	if err != nil {
		exit(fmt.Sprint(err) + ": " + string(port))
	}

	socket := strings.Join([]string{instanceName, strings.TrimSuffix(string(port), "\n")}, ":")
	service := strings.Join([]string{"http", socket}, "://")
	fmt.Printf("[INFO] Service: %s\n", service)

	testPath := strings.Join([]string{"/workdir", testDir, "assert.json"}, "/")
	data, err := ioutil.ReadFile(testPath)
	if err != nil {
		exit(fmt.Sprint(err) + ": " + string(testPath))
	}

	var suite Suite
	err = json.Unmarshal(data, &suite)
	if err != nil {
		exit("Could not parse JSON document.")
	}

	for _, test := range suite.Tests {
		fmt.Println("Endpoint test:", test)
		reqBody, err := json.Marshal(test.Body)
		if err != nil {
			exit(fmt.Sprintf("Bad JSON %s\n", test.Body))
		}

		req, err := http.NewRequest(
			test.Action,
			strings.Join([]string{service, test.Endpoint}, "/"),
			bytes.NewBuffer(reqBody),
		)
		if err != nil {
			exit(fmt.Sprint(err))
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			exit(fmt.Sprintf("Request `%s` failed\n", test.Name))
		}
		defer resp.Body.Close()

		// TODO
	}
}

type Test struct {
	Name     string      `json:"name"`
	Action   string      `json:"action"`
	Endpoint string      `json:"endpoint"`
	Body     interface{} `json:"body"`
	Assert   interface{} `json:"assert"`
}

type Suite struct {
	Tests []Test `json:"tests"`
}

func exit(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}
