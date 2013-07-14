package main

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
)

type LoggerConfig struct {
	GoProcess int
	BindAddr  string
	Configs   []ConfigItem
}

type tempLoggerConfig struct {
}

type ConfigItem struct {
	LoggingName string
	LoggingPath string
	FlushTime   float64
	BufferSize  int
	RotateType  string
}

func LoadJSONConfig(config_path string) (*LoggerConfig, error) {
	// 

	var items []interface{}
	raw_config := make(map[string]interface{})
	logconfig := &LoggerConfig{GO_PROCESS, ADDR, make([]ConfigItem, 0)}

	config_file, err := os.Open(config_path)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(config_file)
	err = dec.Decode(&raw_config)
	if err != nil {
		return nil, err
	}

	// FIXME: to complicated.
	v, ok := raw_config["addr"]
	if !ok {
		return nil, errors.New("Config lack `addr` item.")
	}
	if str_v, ok := v.(string); ok {
		logconfig.BindAddr = str_v
	} else {
		return nil, errors.New("Config `addr` is not string type.")
	}
	v, ok = raw_config["process"]
	if !ok {
		return nil, errors.New("Config lack `process` item.")
	}
	if int_v, ok := v.(float64); ok {
		logconfig.GoProcess = int(int_v)
	} else {
		return nil, errors.New("Config `process` is not integet type.")
	}
	v, ok = raw_config["loggers"]
	if !ok {
		return nil, errors.New("Config lack `loggers` item.")
	}
	if slice_v, ok := v.([]interface{}); ok {
		items = slice_v
	} else {
		return nil, errors.New("Config `loggers` is not list type.")
	}

	for _, v := range items {
		var map_v map[string]interface{}
		var ok bool
		if map_v, ok = v.(map[string]interface{}); ok {
			logconfig.Configs = append(logconfig.Configs, newConfigItemFromMap(map_v))
		} else {
			return nil, errors.New("Config `loggers`.XX is not map type.")
		}
	}

	return logconfig, err
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
				c.LoggingName = strings.ToLower(sv)
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
