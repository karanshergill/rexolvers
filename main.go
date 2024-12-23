package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"

	// YAML Parser
	"gopkg.in/yaml.v3"
)

// Config structure to map YAML
type Config struct {
	SourceURLs []string `yaml:"sourceURLs"`
}

func readSourceURLs(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open YAML file: %v", err)
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("could not decode YAML content: %v", err)
	}

	return config.SourceURLs, nil
}

func fetchURLContent(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status %d for URL %s", resp.StatusCode, url)
	}

	var lines []string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading response body from URL %s: %v", url, err)
	}

	return lines, nil
}

func main() {
	sourcesFile := "config.yaml"

	urls, err := readSourceURLs(sourcesFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	uniqueValues := make(map[string]struct{})

	for _, url := range urls {
		fmt.Printf("Fetching content from: %s\n", url)
		lines, err := fetchURLContent(url)
		if err != nil {
			fmt.Printf("Error fetching URL: %v\n", err)
			continue
		}

		for _, line := range lines {
			uniqueValues[line] = struct{}{}
		}
	}

	outputFile := "resolvers.txt"
	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for value := range uniqueValues {
		_, err := writer.WriteString(value + "\n")
		if err != nil {
			fmt.Printf("Error writing to file: %v\n", err)
			return
		}
	}
	writer.Flush()

	fmt.Printf("Unique values saved to %s\n", outputFile)
}
