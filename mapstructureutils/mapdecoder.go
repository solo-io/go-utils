package mapdecoder

import (
	"encoding/json"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

// NormalizeMapDecode decodes a map[string]interface{} into the given result object
// and handles converting json.Number to int64 or float64 which would cause errors in structpb.NewValue
func NormalizeMapDecode(input interface{}, result interface{}) error {
	config := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(jsonNumberToNumberHook()),
		Result:     result,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

// jsonNumberToNumberHook creates a DecodeHookFuncType that converts json.Number to int64 or float64
func jsonNumberToNumberHook() mapstructure.DecodeHookFuncType {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if numberStr, ok := data.(json.Number); ok {
			if t.Kind() == reflect.Int64 {
				return numberStr.Int64()
			}

			if t.Kind() == reflect.Float64 {
				return numberStr.Float64()
			}
		}

		if f.Kind() == reflect.Map {
			// Recursively process the map
			return convertNumbersInMap(data)
		}

		return data, nil
	}
}

// convertNumbersInMap takes a map and converts all json.Number values to the appropriate numeric type
func convertNumbersInMap(original interface{}) (interface{}, error) {
	resultMap := make(map[string]interface{})
	for key, val := range original.(map[string]interface{}) {
		switch v := val.(type) {
		case json.Number:
			if intVal, err := v.Int64(); err == nil {
				resultMap[key] = intVal
			} else if floatVal, err := v.Float64(); err == nil {
				resultMap[key] = floatVal
			} else {
				// If it's not a number, just keep the original string
				resultMap[key] = val
			}
		case map[string]interface{}:
			// Recursively convert nested maps
			convertedMap, err := convertNumbersInMap(v)
			if err != nil {
				return nil, err
			}
			resultMap[key] = convertedMap
		default:
			resultMap[key] = val
		}
	}
	return resultMap, nil
}
