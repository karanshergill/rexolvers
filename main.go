package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

// config struct to hold source urls
type Config struct {
	PublicSourceURLs  []string `yaml:"publicSourceURLs"`
	TrustedSourceURLs []string `yaml:"trustedSourceURLs"`
}

// cross platform compatibility function to manage the sources file
func getConfigPath() string {
	var configPath string
	if runtime.GOOS == "windows" {
		configPath = filepath.Join(os.Getenv("APPDATA"), "getresolvers", "config.yaml")
	} else {
		configPath = os.ExpandEnv("$HOME/.config/getresolvers/config.yaml")
	}

	// ensure the config directory exists
	configDir := filepath.Dir(configPath)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		os.MkdirAll(configDir, 0755)
	}

	// ensure the config file exists if it doesn't, copy the example config file
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// example config file path
		sampleConfigPath := "config.yaml"
		if _, err := os.Stat(sampleConfigPath); err == nil {
			sourceFile, err := os.Open(sampleConfigPath)
			if err == nil {
				destFile, err := os.Create(configPath)
				if err == nil {
					_, err = io.Copy(destFile, sourceFile)
					sourceFile.Close()
					destFile.Close()
					if err == nil {
						fmt.Printf("Moved sample config file to %s\n", configPath)
					} else {
						fmt.Printf("Error copying sample config file: %v\n", err)
					}
				} else {
					fmt.Printf("Error creating destination config file: %v\n", err)
				}
			} else {
				fmt.Printf("Error opening sample config file: %v\n", err)
			}
		} else {
			fmt.Println("Sample config file not found, skipping copy.")
		}
	}

	return configPath
}

// read source urls from config.yaml
func readSourceURLs(filePath string) (*Config, error) {
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

	return &config, nil
}

// fetch data from the source urls and return a slice of strings
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

// write the unique values to a text file
func saveResolversToFile(uniqueValues map[string]struct{}, outputFile string) error {
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

// process the resolvers from the source urls
func processResolvers(urls []string, outputFile string) error {
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

	return saveResolversToFile(uniqueValues, outputFile)
}

func main() {
	// cli flags
	publicFlag := flag.Bool("public", false, "Process public resolvers")
	trustedFlag := flag.Bool("trusted", false, "Process trusted resolvers")
	allFlag := flag.Bool("all", false, "Process both public and trusted resolvers")
	flag.Parse()

	sourcesFile := getConfigPath()

	// read configuration
	config, err := readSourceURLs(sourcesFile)
	if err != nil {
		fmt.Printf("Error reading configuration: %v\n", err)
		return
	}

	if *allFlag {
		fmt.Println("Processing all resolvers...")
		if err := processResolvers(config.PublicSourceURLs, "public_resolvers.txt"); err != nil {
			fmt.Printf("Error processing public resolvers: %v\n", err)
		}
		if err := processResolvers(config.TrustedSourceURLs, "trusted_resolvers.txt"); err != nil {
			fmt.Printf("Error processing trusted resolvers: %v\n", err)
		}
	} else if *publicFlag {
		fmt.Println("Processing public resolvers...")
		if err := processResolvers(config.PublicSourceURLs, "public_resolvers.txt"); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	} else if *trustedFlag {
		fmt.Println("Processing trusted resolvers...")
		if err := processResolvers(config.TrustedSourceURLs, "trusted_resolvers.txt"); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	} else {
		fmt.Println("Please specify either --public, --trusted, or --all.")
	}
}
