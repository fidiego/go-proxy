package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

//
// Configuration from Env Vars
//

func get_bool_env(name string, default_value bool) bool {
	// load env var and try to get a boolean from it.
	// 1. if the value does not exist, return the default value.
	// 2. elif the value is "true", return true.
	// 3. else, return the default value.
	raw_value, present := os.LookupEnv(name)
	if !present {
		log.Printf("No value found for %s, falling back to default: %v", name, default_value)
		return default_value
	}
	return strings.ToLower(raw_value) == "true"
}

func get_int_env(name string, default_value int) int {
	// load env var and try to get an integer from it.
	// 1. if the value does not exist, return the default value.
	// 2. if the value is a valid int, return the value.
	//    otherwise, log the error and return the default value.
	raw_value, present := os.LookupEnv(name)
	if !present {
		log.Printf("No value found for %s, falling back to default: %v", name, default_value)
		return default_value
	}
	parsed, err := strconv.Atoi(raw_value)
	if err != nil {
		log.Fatal("Invalid value found for %s, falling back to default: %v", name, default_value)
	}
	return parsed
}

func get_string_env(name string, default_value string) string {
	// load env var and try to cast to string.
	// if the value does not exist, return the default value.
	raw_value := os.Getenv(name)
	if len(raw_value) == 0 {
		log.Printf("No value found for %s, falling back to default, '%s'", name, default_value)
		return default_value
	} else {
		return raw_value
	}
}

type Configs struct {
	// sub-items
	BaseUrl     string
	ServiceName string
	Https       bool
	Port        int

	Environment string
	Debug       bool // determines verbosity of logger

	// secret key
	SecretKey string
}

func (c *Configs) Load() {
	log.Printf("Loading Configs")
	c.BaseUrl = get_string_env("BASE_URL", "localhost")
	c.ServiceName = get_string_env("SERVICE_NAME", "Go-Proxy")
	c.Https = get_bool_env("HTTPS", false)
	c.Port = get_int_env("PORT", 8080)
	c.Environment = get_string_env("ENVIRONMENT", "production")
	c.Debug = get_bool_env("DEBUG", false)
}
