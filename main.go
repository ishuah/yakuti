package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

type Request struct {
	Path        string
	ClientIP    string
	StartedAt   string
	CompletedAt string
	Method      string
	URLParams   string
	Duration    int
	Status      string
}

var Requests map[string]Request

type Path struct {
	Count              int
	CumulativeDuration int
}

var Paths map[string]Path

func main() {

	filePath := os.Args[1]
	file, err := os.Open(filePath)

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	Requests = make(map[string]Request)
	Paths = make(map[string]Path)

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	startRegex := regexp.MustCompile(`\[(?P<timestamp>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+).*INFO -- : \[(?P<requestID>\w{8}-\w{4}-\w{4}-\w{4}-\w{12})] Started (?P<method>\w+) \"(?P<endpoint>\/*(\w+\/*\-*\?*\=*\.*\&*)*?)\" for (?P<ipaddress>\d{1,3}.\d{1,3}.\d{1,3}.\d{1,3}).*`)
	completedRegex := regexp.MustCompile(`\[(?P<timestamp>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d+).*INFO -- : \[(?P<requestID>\w{8}-\w{4}-\w{4}-\w{4}-\w{12})] Completed (?P<status>\d{3}) [\w\s]+ in (?P<duration>\d+)ms`)

	for scanner.Scan() {
		if startRegex.MatchString(scanner.Text()) {
			tokens := startRegex.FindAllStringSubmatch(scanner.Text(), -1)
			startedAt := tokens[0][1]
			requestID := tokens[0][2]
			method := tokens[0][3]
			path := tokens[0][4]
			clientIP := tokens[0][6]

			request := Request{Path: path, ClientIP: clientIP, StartedAt: startedAt, Method: method}

			if strings.Contains(path, "?") {
				split := strings.Split(path, "?")
				request.Path = split[0]
				request.URLParams = split[1]
			}

			Requests[requestID] = request

		} else if completedRegex.MatchString(scanner.Text()) {
			tokens := completedRegex.FindAllStringSubmatch(scanner.Text(), -1)
			completedAt := tokens[0][1]
			requestID := tokens[0][2]
			status := tokens[0][3]
			duration, _ := strconv.Atoi(tokens[0][4])

			request := Requests[requestID]
			request.CompletedAt = completedAt
			request.Status = status
			request.Duration = duration
			Requests[requestID] = request

			if request.Path == "" {
				fmt.Println(requestID)
			}

			if _, ok := Paths[request.Path]; !ok {
				Paths[request.Path] = Path{Count: 0, CumulativeDuration: 0}
			}

			path := Paths[request.Path]
			path.Count++
			path.CumulativeDuration += request.Duration
			Paths[request.Path] = path
		}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Path", "Request Count", "Average Response Time"})

	for endpoint, path := range Paths {
		avgDuration := float64(path.CumulativeDuration) / float64(path.Count)
		table.Append([]string{endpoint, fmt.Sprintf("%d", path.Count), fmt.Sprintf("%f ms", avgDuration)})
	}
	table.Render()
}
