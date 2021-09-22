package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"sync"
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
	fmt.Printf("[INFO] Service endpoint: %s\n", service)

	testPath := strings.Join([]string{"/workdir", testDir, "assert.json"}, "/")
	data, err := ioutil.ReadFile(testPath)
	if err != nil {
		exit(fmt.Sprint(err) + ": " + string(testPath))
	}

	var tests []Test
	err = json.Unmarshal(data, &tests)
	if err != nil {
		exit(fmt.Sprint(err))
	}

	suite := New(tests)

	var wg sync.WaitGroup
	wg.Add(suite.Total)

	for _, test := range suite.Tests {
		go func(test Test) {
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

			var res map[string]interface{}
			decoder := json.NewDecoder(resp.Body)
			err = decoder.Decode(&res)
			if err != nil {
				exit(fmt.Sprintf("Could not decode raw JSON in the response: %s", resp.Body))
			}

			suite.Assert(test, res)
			wg.Done()
		}(test)
	}

	wg.Wait()
	suite.Log()
}

type Test struct {
	Name     string      `json:"name"`
	Action   string      `json:"action"`
	Endpoint string      `json:"endpoint"`
	Body     interface{} `json:"body"`
	Assert   interface{} `json:"assert"`
}

type Suite struct {
	Total  int
	Passed int
	Tests  []Test `json:"tests"`
}

func New(tests []Test) Suite {
	return Suite{
		Total:  len(tests),
		Passed: 0,
		Tests:  tests,
	}
}

func (s *Suite) Assert(test Test, res map[string]interface{}) {
	if reflect.DeepEqual(test.Assert, res) {
		s.Passed++
		fmt.Println(fmt.Sprintf("%s...passed", test.Name))
	} else {
		fmt.Println(fmt.Sprintf("%s...failed", test.Name))
	}
}

func (s Suite) Log() {
	fmt.Println("[INFO] Test results:")
	fmt.Println("Total tests: ", s.Total)
	fmt.Println(fmt.Sprintf("\t%d...passed", s.Passed))
	fmt.Println(fmt.Sprintf("\t%d...failed", s.Total-s.Passed))
	if s.Passed != s.Total {
		exit("One or more tests failed, exiting...")
	}
}

func exit(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}
