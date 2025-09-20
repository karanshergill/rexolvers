package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v3"
)

// config struct to hold source urls
type Config struct {
	PublicSourceURLs  []string `yaml:"publicSourceURLs"`
	TrustedSourceURLs []string `yaml:"trustedSourceURLs"`
}

// database functions
func initDatabase() (*sql.DB, error) {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./resolvers.db"
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS resolvers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ip_address TEXT UNIQUE NOT NULL,
		resolver_type TEXT NOT NULL,
		source_url TEXT,
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, fmt.Errorf("error creating table: %v", err)
	}

	return db, nil
}

func saveResolverToDB(db *sql.DB, ipAddress, resolverType, sourceURL string) error {
	insertSQL := `INSERT OR IGNORE INTO resolvers (ip_address, resolver_type, source_url) VALUES (?, ?, ?)`
	_, err := db.Exec(insertSQL, ipAddress, resolverType, sourceURL)
	return err
}

func getResolversFromDB(db *sql.DB, resolverType string) ([]string, error) {
	var query string
	var args []interface{}

	if resolverType == "all" {
		query = "SELECT ip_address FROM resolvers ORDER BY ip_address"
	} else {
		query = "SELECT ip_address FROM resolvers WHERE resolver_type = ? ORDER BY ip_address"
		args = []interface{}{resolverType}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resolvers []string
	for rows.Next() {
		var ip string
		if err := rows.Scan(&ip); err != nil {
			return nil, err
		}
		resolvers = append(resolvers, ip)
	}

	return resolvers, nil
}

func getResolverStats(db *sql.DB) error {
	query := "SELECT resolver_type, COUNT(*) FROM resolvers GROUP BY resolver_type"
	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	fmt.Println("Database Statistics:")
	total := 0
	for rows.Next() {
		var resolverType string
		var count int
		if err := rows.Scan(&resolverType, &count); err != nil {
			return err
		}
		fmt.Printf("  %s: %d resolvers\n", resolverType, count)
		total += count
	}
	fmt.Printf("  total: %d resolvers\n", total)
	return nil
}

func clearResolversByType(db *sql.DB, resolverType string) error {
	deleteSQL := "DELETE FROM resolvers WHERE resolver_type = ?"
	result, err := db.Exec(deleteSQL, resolverType)
	if err != nil {
		return fmt.Errorf("error clearing %s resolvers: %v", resolverType, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	fmt.Printf("Cleared %d existing %s resolvers from database\n", rowsAffected, resolverType)
	return nil
}

// cross platform compatibility function to manage the sources file
func getConfigPath() string {
	var configPath string
	if runtime.GOOS == "windows" {
		configPath = filepath.Join(os.Getenv("APPDATA"), "rexolvers", "config.yaml")
	} else {
		configPath = os.ExpandEnv("$HOME/.config/rexolvers/config.yaml")
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
func processResolvers(urls []string, outputFile string, resolverType string, db *sql.DB, saveToFile bool) error {
	// Clear existing public resolvers from database before processing new ones
	if db != nil && resolverType == "public" {
		if err := clearResolversByType(db, "public"); err != nil {
			fmt.Printf("Warning: %v\n", err)
		}
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

			// Save to database if db connection provided
			if db != nil {
				if err := saveResolverToDB(db, line, resolverType, url); err != nil {
					// Only show warning for trusted resolvers (duplicates expected for public after clearing)
					if resolverType == "trusted" {
						fmt.Printf("Warning: Could not save resolver %s to database: %v\n", line, err)
					}
				}
			}
		}
	}

	// Save to file if requested
	if saveToFile {
		return saveResolversToFile(uniqueValues, outputFile)
	}

	fmt.Printf("Processed %d unique %s resolvers\n", len(uniqueValues), resolverType)
	return nil
}

func main() {
	// cli flags
	publicFlag := flag.Bool("public", false, "Process public resolvers")
	trustedFlag := flag.Bool("trusted", false, "Process trusted resolvers")
	allFlag := flag.Bool("all", false, "Process both public and trusted resolvers")
	dbFlag := flag.Bool("db", false, "Save resolvers to database")
	fileFlag := flag.Bool("file", true, "Save resolvers to file (default: true)")
	listFlag := flag.String("list", "", "List resolvers from database (public|trusted|all)")
	statsFlag := flag.Bool("stats", false, "Show database statistics")
	flag.Parse()

	// Handle list operation first
	if *listFlag != "" {
		db, err := initDatabase()
		if err != nil {
			fmt.Printf("Error initializing database: %v\n", err)
			return
		}
		defer db.Close()

		resolvers, err := getResolversFromDB(db, *listFlag)
		if err != nil {
			fmt.Printf("Error retrieving resolvers from database: %v\n", err)
			return
		}

		fmt.Printf("Found %d %s resolvers in database:\n", len(resolvers), *listFlag)
		for _, resolver := range resolvers {
			fmt.Println(resolver)
		}
		return
	}

	// Handle stats operation
	if *statsFlag {
		db, err := initDatabase()
		if err != nil {
			fmt.Printf("Error initializing database: %v\n", err)
			return
		}
		defer db.Close()

		if err := getResolverStats(db); err != nil {
			fmt.Printf("Error retrieving statistics: %v\n", err)
			return
		}
		return
	}

	sourcesFile := getConfigPath()

	// read configuration
	config, err := readSourceURLs(sourcesFile)
	if err != nil {
		fmt.Printf("Error reading configuration: %v\n", err)
		return
	}

	// Initialize database if needed
	var db *sql.DB
	if *dbFlag {
		db, err = initDatabase()
		if err != nil {
			fmt.Printf("Error initializing database: %v\n", err)
			return
		}
		defer db.Close()
		fmt.Println("Database initialized successfully")
	}

	if *allFlag {
		fmt.Println("Processing all resolvers...")
		if err := processResolvers(config.PublicSourceURLs, "public_resolvers.txt", "public", db, *fileFlag); err != nil {
			fmt.Printf("Error processing public resolvers: %v\n", err)
		}
		if err := processResolvers(config.TrustedSourceURLs, "trusted_resolvers.txt", "trusted", db, *fileFlag); err != nil {
			fmt.Printf("Error processing trusted resolvers: %v\n", err)
		}
	} else if *publicFlag {
		fmt.Println("Processing public resolvers...")
		if err := processResolvers(config.PublicSourceURLs, "public_resolvers.txt", "public", db, *fileFlag); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	} else if *trustedFlag {
		fmt.Println("Processing trusted resolvers...")
		if err := processResolvers(config.TrustedSourceURLs, "trusted_resolvers.txt", "trusted", db, *fileFlag); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	} else {
		fmt.Println("Usage:")
		fmt.Println("  Process resolvers: --public, --trusted, or --all")
		fmt.Println("  List from database: --list=public|trusted|all")
		fmt.Println("  Show statistics: --stats")
		fmt.Println("  Optional: --db (save to database), --file=false (skip file output)")
	}
}
