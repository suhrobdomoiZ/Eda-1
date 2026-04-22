package config

import (
	"fmt"
	"os"
)

type Key string

func (key Key) MustGet() string {
	val := os.Getenv(string(key))

	if val == "" {
		panic(fmt.Sprintf("config.MustGet: required value %s is not set", string(key)))
	}
	return val

}

func (key Key) Get(defaultValue string) string {
	if val := os.Getenv(string(key)); val != "" {
		return val
	}
	return defaultValue
}
