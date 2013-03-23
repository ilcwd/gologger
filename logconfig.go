package main

import (
	"encoding/json"
	"os"
)

type ConfigItem struct {
	LoggingName string
	LoggingPath string
	FlushTime   float64
	BufferSize  int
	RotateType  string
}

func LoadJSONConfig(config_path string) ([]ConfigItem, error) {
	items := make([]map[string]interface{}, 100)
	result := make([]ConfigItem, 0)
	config_file, err := os.Open(config_path)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(config_file)
	err = dec.Decode(&items)
	if err != nil {
		return nil, err
	}

	for _, v := range items {
		result = append(result, newConfigItemFromMap(v))
	}

	return result, err
}

func newConfigItemFromMap(m map[string]interface{}) ConfigItem {
	c := ConfigItem{
		LOGGING_NAME,
		LOGGING_PATH,
		FLUSH_TIME,
		BUFFER_SIZE,
		ROTATE_TYPE,
	}
	for k, v := range m {
		switch k {
		case "LoggingName":
			if sv, ok := v.(string); ok {
				c.LoggingName = sv
			}
		case "LoggingPath":
			if sv, ok := v.(string); ok {
				c.LoggingPath = sv
			}
		case "FlushTime":
			if sv, ok := v.(float64); ok {
				c.FlushTime = sv
			}
		case "BufferSize":
			if sv, ok := v.(float64); ok {
				c.BufferSize = int(sv)
			}
		case "RotateType":
			if sv, ok := v.(string); ok {
				c.RotateType = sv
			}

		}
	}
	return c
}
