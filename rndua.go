package main

import (
	"crypto/rand"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	defaultURL = "https://cdn.jsdelivr.net/gh/microlinkhq/top-user-agents@master/src/index.json"
	cacheDir   = ".cache/useragent-cli"
	cacheFile  = "cache.json"
	cacheTTL   = 24 * time.Hour
)

type CacheData struct {
	UserAgents []string  `json:"user_agents"`
	Timestamp  time.Time `json:"timestamp"`
}

type Config struct {
	Count   int
	Format  string
	Refresh bool
}

// Embedded backup user agents
var backupUserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/120.0",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (iPad; CPU OS 17_2 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15",
	"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Android 14; Mobile; rv:109.0) Gecko/121.0 Firefox/121.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/119.0",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 17_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/120.0",
	"Mozilla/5.0 (Windows NT 11.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Android 13; Mobile; rv:109.0) Gecko/121.0 Firefox/121.0",
	"Mozilla/5.0 (iPad; CPU OS 17_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/119.0.0.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/120.0",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 16_7 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/118.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Safari/605.1.15",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Android 12; Mobile; rv:109.0) Gecko/121.0 Firefox/121.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36",
	"Mozilla/5.0 (iPad; CPU OS 16_7 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/119.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Android 11; Mobile; rv:109.0) Gecko/121.0 Firefox/121.0",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 16_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/119.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/117.0",
	"Mozilla/5.0 (Android 10; Mobile; rv:109.0) Gecko/121.0 Firefox/121.0",
	"Mozilla/5.0 (iPad; CPU OS 16_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 15_7 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.6 Mobile/15E148 Safari/604.1",
}

func main() {
	var config Config
	
	flag.IntVar(&config.Count, "count", 1, "Number of user agents to output")
	flag.StringVar(&config.Format, "format", "txt", "Output format: txt, json, csv")
	flag.BoolVar(&config.Refresh, "refresh", false, "Force refresh the cache")
	flag.Parse()

	// Validate format
	if config.Format != "txt" && config.Format != "json" && config.Format != "csv" {
		log.Fatal("Invalid format. Use: txt, json, or csv")
	}

	// Validate count
	if config.Count < 1 {
		log.Fatal("Count must be at least 1")
	}

	// Get user agents
	userAgents, err := getUserAgents(config.Refresh)
	if err != nil {
		log.Fatalf("Failed to get user agents: %v", err)
	}

	if len(userAgents) == 0 {
		log.Fatal("No user agents available")
	}

	// Select random user agents
	selected, err := selectRandomUserAgents(userAgents, config.Count)
	if err != nil {
		log.Fatalf("Failed to select random user agents: %v", err)
	}

	// Output in requested format
	if err := outputUserAgents(selected, config.Format); err != nil {
		log.Fatalf("Failed to output user agents: %v", err)
	}
}

func getUserAgents(forceRefresh bool) ([]string, error) {
	// Try to get from cache first (unless force refresh)
	if !forceRefresh {
		if userAgents, err := getUserAgentsFromCache(); err == nil && len(userAgents) > 0 {
			return userAgents, nil
		}
	}

	// Fetch from URL
	userAgents, err := fetchUserAgentsFromURL(defaultURL)
	if err == nil && len(userAgents) > 0 {
		// Cache the result
		if err := cacheUserAgents(userAgents); err != nil {
			log.Printf("Warning: Failed to cache user agents: %v", err)
		}
		return userAgents, nil
	}

	// If URL fetch fails, try cache as last resort
	if userAgents, cacheErr := getUserAgentsFromCache(); cacheErr == nil && len(userAgents) > 0 {
		log.Printf("Warning: Failed to fetch from URL (%v), using cached data", err)
		return userAgents, nil
	}

	// Finally, use embedded backup
	log.Printf("Warning: Failed to fetch from URL (%v), using backup data", err)
	return backupUserAgents, nil
}

func fetchUserAgentsFromURL(url string) ([]string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// The jsdelivr URL returns JSON array of strings
	var userAgents []string
	if err := json.Unmarshal(body, &userAgents); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return userAgents, nil
}

func getCacheFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	
	cacheDirPath := filepath.Join(homeDir, cacheDir)
	if err := os.MkdirAll(cacheDirPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}
	
	return filepath.Join(cacheDirPath, cacheFile), nil
}

func getUserAgentsFromCache() ([]string, error) {
	cacheFilePath, err := getCacheFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cacheFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var cache CacheData
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("failed to parse cache file: %w", err)
	}

	// Check if cache is still valid
	if time.Since(cache.Timestamp) > cacheTTL {
		return nil, fmt.Errorf("cache expired")
	}

	return cache.UserAgents, nil
}

func cacheUserAgents(userAgents []string) error {
	cacheFilePath, err := getCacheFilePath()
	if err != nil {
		return err
	}

	cache := CacheData{
		UserAgents: userAgents,
		Timestamp:  time.Now(),
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	if err := os.WriteFile(cacheFilePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

func selectRandomUserAgents(userAgents []string, count int) ([]string, error) {
	if count >= len(userAgents) {
		// If we need more than available, return all (with potential duplicates)
		result := make([]string, count)
		for i := 0; i < count; i++ {
			idx, err := cryptoRandInt(len(userAgents))
			if err != nil {
				return nil, fmt.Errorf("failed to generate random number: %w", err)
			}
			result[i] = userAgents[idx]
		}
		return result, nil
	}

	// Select unique random user agents
	selected := make([]string, 0, count)
	used := make(map[int]bool)

	for len(selected) < count {
		idx, err := cryptoRandInt(len(userAgents))
		if err != nil {
			return nil, fmt.Errorf("failed to generate random number: %w", err)
		}

		if !used[idx] {
			used[idx] = true
			selected = append(selected, userAgents[idx])
		}
	}

	return selected, nil
}

func cryptoRandInt(max int) (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(n.Int64()), nil
}

func outputUserAgents(userAgents []string, format string) error {
	switch format {
	case "txt":
		for _, ua := range userAgents {
			fmt.Println(ua)
		}
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(userAgents)
	case "csv":
		writer := csv.NewWriter(os.Stdout)
		defer writer.Flush()
		
		for _, ua := range userAgents {
			if err := writer.Write([]string{ua}); err != nil {
				return fmt.Errorf("failed to write CSV record: %w", err)
			}
		}
	}
	return nil
}