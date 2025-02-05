package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// config struct to hold source urls
type Config struct {
	SourceURLs []string `yaml:"sourceURLs"`
}

// read source urls from a config.yaml file
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

// fetch data from a url and return a slice of strings
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

func processPublic(sourcesFile, outputFile string) error {
	urls, err := readSourceURLs(sourcesFile)
	if err != nil {
		return fmt.Errorf("error reading source URLs: %v", err)
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

	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for value := range uniqueValues {
		_, err := writer.WriteString(value + "\n")
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
	}
	writer.Flush()

	fmt.Printf("Unique values saved to %s\n", outputFile)
	return nil
}

func processTrusted(sourcesFile, outputFile string) error {
	// Placeholder for processing trusted resolvers logic
	fmt.Println("Processing trusted resolvers...")
	// Future logic here
	fmt.Println("Trusted resolvers processed (placeholder).")
	return nil
}

func main() {
	// cli flags
	publicFlag := flag.Bool("public", false, "Process public resolvers")
	trustedFlag := flag.Bool("trusted", false, "Process trusted resolvers")
	allFlag := flag.Bool("all", false, "Process both public and trusted resolvers")
	flag.Parse()

	sourcesFile := "config.yaml"

	if *allFlag {
		fmt.Println("Processing all resolvers...")
		err := processPublic(sourcesFile, "public_resolvers.txt")
		if err != nil {
			fmt.Printf("Error processing public resolvers: %v\n", err)
		}
		err = processTrusted(sourcesFile, "trusted_resolvers.txt")
		if err != nil {
			fmt.Printf("Error processing trusted resolvers: %v\n", err)
		}
	} else if *publicFlag {
		fmt.Println("Processing public resolvers...")
		err := processPublic(sourcesFile, "public_resolvers.txt")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	} else if *trustedFlag {
		fmt.Println("Processing trusted resolvers...")
		err := processTrusted(sourcesFile, "trusted_resolvers.txt")
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	} else {
		fmt.Println("Please specify either --public, --trusted, or --all.")
	}
}
