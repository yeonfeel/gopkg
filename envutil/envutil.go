package envutil

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// SetEnvB sets a variable from env
func SetEnvB(envVar *bool, envName string) {
	if v := os.Getenv(envName); v != "" {
		if v == "true" {
			*envVar = true
		} else {
			*envVar = false
		}
	}
}

// SetEnvS sets a vriable from env
func SetEnvS(envVar *string, envName string) {
	if v := os.Getenv(envName); v != "" {
		*envVar = v
	}
}

// SetEnvA sets a vriable from env
func SetEnvA(envVar *[]string, envName string, separator string) {
	if v := os.Getenv(envName); v != "" {
		*envVar = strings.Split(v, separator)
	}
}

// SetEnvI sets a vriable from env
func SetEnvI(envVar *int, envName string) {
	if v := os.Getenv(envName); v != "" {
		var err error
		*envVar, err = strconv.Atoi(v)
		if err != nil {
			panic(fmt.Sprintf("Env[%s] is not int: [%s]", envName, v))
		}
	}
}

// SetEnvI64 sets a vriable from env
func SetEnvI64(envVar *int64, envName string) {
	if v := os.Getenv(envName); v != "" {
		var err error
		*envVar, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			panic(fmt.Sprintf("Env[%s] is not int64: [%s]", envName, v))
		}
	}
}

// SetEnvT sets a vriable from env
func SetEnvT(envVar *time.Duration, envName string) {
	if v := os.Getenv(envName); v != "" {
		var err error
		*envVar, err = time.ParseDuration(v)
		if err != nil {
			panic(fmt.Sprintf("Env[%s] is not duration: %s", envName, v))
		}
	}
}

// SetEnv sets a vriable from env
func SetEnv(envVar interface{}, envName string) {
	if v := os.Getenv(envName); v != "" {
		var err error
		err = json.Unmarshal([]byte(v), envVar)
		if err != nil {
			panic(fmt.Sprintf("Env[%s] is not %s: %s", envName, reflect.ValueOf(envVar).Type(), v))
		}
	}
}
