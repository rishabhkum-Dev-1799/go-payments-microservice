package config

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Env read typed values from the environment variables and also captures the problem, it doesnt stop at the first problem it captures all the problems and return all problems combined
type Env struct {
	lookup   func(key string) (string, error)
	problems []string
}

func New(lookup func(key string) (string, error)) *Env {
	return &Env{lookup: lookup}
}

// get returns the value of the key passed in the function by trimming it ...r
func (e *Env) get(key string) (string, bool) {
	v, _ := e.lookup(key)
	v = strings.TrimSpace(v)
	return v, v != ""
}

// Invalid function records the problem with the variable
func (e *Env) Invalid(key string, problem string) {
	e.problems = append(e.problems, key+": "+problem)
}

// Err returns the combined errors ...r
func (e *Env) Err() error {
	if len(e.problems) == 0 {
		return nil
	}
	return errors.New("invalid configuration\n:" + strings.Join(e.problems, "\n "))
}

// string is the function which will return the value of the key and if unset then default one
func (e *Env) String(key string, def string) string {
	if v, ok := e.get(key); ok {
		return v
	}
	return def
}

// Required : Required is the function which will extract the important requried parameters or else logs the erpor
func (e *Env) Required(key string) string {
	v, ok := e.get(key)
	if !ok {
		e.Invalid(key, "the key is required but not set")
	}
	return v
}

// One of : one of is the function that will the check the key should have values from only the required options
func (e *Env) Oneof(key string, def string, allowed ...string) string {
	v := e.String(key, def)
	if !slices.Contains(allowed, v) {
		e.Invalid(key, fmt.Sprintf("%q is not present in the allowed values:%s", v, strings.Join(allowed, ", ")))
	}
	return v
}

// Parse is the generic function that will convert the key value into the particular type if it doesnt parse return the default value
func Parse[T any](e *Env, key string, def T, parse func(key string) (T, error)) T {
	v, ok := e.get(key)
	if !ok {
		return def
	}
	x, err := parse(v)
	if err != nil {
		e.Invalid(key, fmt.Sprintf("invalid value:%q", v))
		return def
	}
	return x
}

// Int : Int is the function that will the convert the given property key from string to Int format ...r
func (e *Env) Int(key string, def int) int {
	return Parse(e, key, def, strconv.Atoi)
}

// Duration :Duration is the function that will return in the Duration fomat or def value ...r
func (e *Env) Duration(key string, def time.Duration) time.Duration {
	return Parse(e, key, def, time.ParseDuration)
}
