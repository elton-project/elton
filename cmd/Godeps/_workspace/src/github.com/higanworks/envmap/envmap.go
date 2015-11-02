// Create map object of environment variables.
package envmap

import (
	"os"
	"regexp"
	"strings"
)

var osEnv = func() []string {
	return os.Environ()
}

// Returns all environment variables as key-value map.
func All() map[string]string {
	data := osEnv()
	items := make(map[string]string)
	for _, val := range data {
		splits := strings.SplitN(val, "=", 2)
		key := splits[0]
		value := splits[1]
		items[key] = value
	}
	return items
}

// Returns filtered by matched keys of environment variables as key-value.
func Matched(rule string) map[string]string {
	data := osEnv()
	items := make(map[string]string)
	for _, val := range data {
		splits := strings.SplitN(val, "=", 2)
		matched_flag, _ := regexp.MatchString(rule, splits[0])
		if matched_flag {
			key := splits[0]
			value := splits[1]
			items[key] = value
		}
	}
	return items
}

// Returns Keys of all environment variables.
func ListKeys() []string {
	data := osEnv()
	var keys []string
	for _, val := range data {
		splits := strings.SplitN(val, "=", 2)
		keys = append(keys, splits[0])
	}
	return keys
}
