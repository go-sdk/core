package config

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/go-sdk/core/errx"
)

var variablePattern = regexp.MustCompile(`\$\{([^{}]+)}`)

func resolveVariables(input map[string]any) (map[string]any, error) {
	resolver := variableResolver{
		input:  input,
		output: make(map[string]any, len(input)),
		state:  make(map[string]uint8, len(input)),
	}
	for key := range input {
		if _, err := resolver.resolveKey(key); err != nil {
			return nil, err
		}
	}
	return resolver.output, nil
}

type variableResolver struct {
	input  map[string]any
	output map[string]any
	state  map[string]uint8
}

func (r *variableResolver) resolveKey(key string) (any, error) {
	switch r.state[key] {
	case 1:
		return nil, errx.Newf("cyclic reference detected at %q", key)
	case 2:
		return r.output[key], nil
	}
	value, ok := r.input[key]
	if !ok {
		return nil, errx.Newf("referenced key %q does not exist", key)
	}
	r.state[key] = 1
	resolved, err := r.resolveValue(value)
	if err != nil {
		return nil, errx.Wrapf(err, "resolve %q", key)
	}
	r.output[key] = resolved
	r.state[key] = 2
	return resolved, nil
}

func (r *variableResolver) resolveValue(value any) (any, error) {
	switch value := value.(type) {
	case string:
		var replaceErr error
		resolved := variablePattern.ReplaceAllStringFunc(value, func(match string) string {
			if replaceErr != nil {
				return match
			}
			parts := variablePattern.FindStringSubmatch(match)
			reference := parts[1]
			referenced, err := r.resolveKey(reference)
			if err != nil {
				replaceErr = err
				return match
			}
			if !isScalar(referenced) {
				replaceErr = errx.Newf("referenced key %q is not a scalar value", reference)
				return match
			}
			return fmt.Sprint(referenced)
		})
		if replaceErr != nil {
			return nil, replaceErr
		}
		return resolved, nil
	case []any:
		resolved := make([]any, len(value))
		for index := range value {
			item, err := r.resolveValue(value[index])
			if err != nil {
				return nil, errx.Wrapf(err, "resolve array index %d", index)
			}
			resolved[index] = item
		}
		return resolved, nil
	case map[string]any:
		resolved := make(map[string]any, len(value))
		for key, item := range value {
			itemValue, err := r.resolveValue(item)
			if err != nil {
				return nil, errx.Wrapf(err, "resolve map key %q", key)
			}
			resolved[key] = itemValue
		}
		return resolved, nil
	default:
		return value, nil
	}
}

func isScalar(value any) bool {
	if value == nil {
		return false
	}
	switch reflect.TypeOf(value).Kind() {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}
