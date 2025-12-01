package main

import (
	"os"
	"strconv"
)

// Configuration holds all configurable parameters for the deep research system
type Configuration struct {
	// Model configurations
	SummarizationModel          string
	SummarizationModelMaxTokens int
	ResearchModel               string
	ResearchModelMaxTokens      int
	CompressionModel            string
	CompressionModelMaxTokens   int
	FinalReportModel            string
	FinalReportModelMaxTokens   int

	// Search API configuration
	SearchAPI    string
	TavilyAPIKey string

	// Research parameters
	MaxResearcherIterations    int
	MaxConcurrentResearchUnits int
	MaxContentLength           int
	MaxToolCallIterations      int
}

// NewConfiguration creates a new Configuration with default values from environment
func NewConfiguration() *Configuration {
	return &Configuration{
		// Model defaults
		SummarizationModel:          getEnvOrDefault("SUMMARIZATION_MODEL", "deepseek-v3"),
		SummarizationModelMaxTokens: getEnvIntOrDefault("SUMMARIZATION_MODEL_MAX_TOKENS", 4096),
		ResearchModel:               getEnvOrDefault("RESEARCH_MODEL", "deepseek-v3"),
		ResearchModelMaxTokens:      getEnvIntOrDefault("RESEARCH_MODEL_MAX_TOKENS", 10000),
		CompressionModel:            getEnvOrDefault("COMPRESSION_MODEL", "deepseek-v3"),
		CompressionModelMaxTokens:   getEnvIntOrDefault("COMPRESSION_MODEL_MAX_TOKENS", 8192),
		FinalReportModel:            getEnvOrDefault("FINAL_REPORT_MODEL", "deepseek-v3"),
		FinalReportModelMaxTokens:   getEnvIntOrDefault("FINAL_REPORT_MODEL_MAX_TOKENS", 10000),

		// Search API defaults
		SearchAPI:    getEnvOrDefault("SEARCH_API", "tavily"),
		TavilyAPIKey: os.Getenv("TAVILY_API_KEY"),

		// Research parameters
		MaxResearcherIterations:    getEnvIntOrDefault("MAX_RESEARCHER_ITERATIONS", 10),
		MaxConcurrentResearchUnits: getEnvIntOrDefault("MAX_CONCURRENT_RESEARCH_UNITS", 3),
		MaxContentLength:           getEnvIntOrDefault("MAX_CONTENT_LENGTH", 50000),
		MaxToolCallIterations:      getEnvIntOrDefault("MAX_TOOL_CALL_ITERATIONS", 20),
	}
}

// Helper functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
