package config

import (
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strings"

	"github.com/hashicorp/go-envparse"
	"gopkg.in/yaml.v3"
)

type ConfigLoader struct {
	filePath string
	envVars  map[string]string
}

type ConfigLoaderOpts struct {
	FilePath string
}

func NewConfigLoader(opts *ConfigLoaderOpts) *ConfigLoader {
	if opts == nil {
		opts = &ConfigLoaderOpts{}
	}

	// If no file path is provided, try to find default config files
	filePath := opts.FilePath
	if filePath == "" {
		filePath = findDefaultConfigFile()
	}

	return &ConfigLoader{
		filePath: opts.FilePath,
		envVars:  make(map[string]string),
	}
}

// findDefaultConfigFile looks for configuration files in the default locations
func findDefaultConfigFile() string {
	// List of default config files to look for (in order of preference)
	defaultFiles := []string{
		"application.yaml",
		"application.yml",
		".env",
	}

	for _, file := range defaultFiles {
		if _, err := os.Stat(file); err == nil {
			// File exists
			return file
		}
	}

	// No default config file found
	return ""
}

func (c *ConfigLoader) Load(cfg any) error {
	// Load file if found
	if c.filePath != "" {
		file, err := c.openFile()
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer func(file *os.File) {
			err = file.Close()
			if err != nil {
				slog.Error(fmt.Sprintf("failed to close file: %s", c.filePath), "error", err)
			}
		}(file)

		// Try YAML first
		yamlOk, yamlErr := c.loadYamlFile(file)
		if yamlOk {
			slog.Info("Loaded configuration from YAML file")
		} else {
			// Fall back to ENV format
			envOk, envErr := c.loadEnvFile(file)
			if envOk {
				slog.Info("Loaded configuration from ENV file")
			} else {
				slog.Info("Failed to load file as YAML or ENV format",
					"yamlError", yamlErr,
					"envError", envErr)
			}
		}
	}

	// Process the loaded environment variables
	return c.processStruct(reflect.ValueOf(cfg).Elem())
}

func (c *ConfigLoader) openFile() (*os.File, error) {
	return os.Open(c.filePath)
}

func (c *ConfigLoader) loadYamlFile(file *os.File) (bool, error) {
	if _, err := file.Seek(0, 0); err != nil {
		return false, fmt.Errorf("failed to reset file position: %w", err)
	}

	data, err := os.ReadFile(c.filePath)
	if err != nil {
		return false, fmt.Errorf("failed to read YAML file: %w", err)
	}

	var yamlMap map[string]interface{}
	if err = yaml.Unmarshal(data, &yamlMap); err != nil {
		return false, fmt.Errorf("failed to parse YAML file: %w", err)
	}

	c.flattenYAML("", yamlMap)
	return true, nil
}

// flattenYAML converts nested YAML structure to flat key-value pairs
func (c *ConfigLoader) flattenYAML(prefix string, m map[string]interface{}) {
	for k, v := range m {
		var newPrefix string
		if prefix == "" {
			newPrefix = strings.ToUpper(k)
		} else {
			newPrefix = prefix + "_" + strings.ToUpper(k)
		}

		switch vt := v.(type) {
		case map[string]interface{}:
			c.flattenYAML(newPrefix, vt)
		case []interface{}:
			// Handle arrays by using index in the key
			for i, item := range vt {
				arrayKey := fmt.Sprintf("%s_%d", newPrefix, i)
				if mapItem, ok := item.(map[string]interface{}); ok {
					c.flattenYAML(arrayKey, mapItem)
				} else {
					c.envVars[arrayKey] = fmt.Sprintf("%v", item)
				}
			}
		default:
			c.envVars[newPrefix] = fmt.Sprintf("%v", vt)
		}
	}
}

func (c *ConfigLoader) loadEnvFile(file *os.File) (bool, error) {
	if _, err := file.Seek(0, 0); err != nil {
		return false, fmt.Errorf("failed to reset file position: %w", err)
	}

	envVars, err := envparse.Parse(file)
	if err != nil {
		return false, err
	}

	c.envVars = envVars
	return true, nil
}

func (c *ConfigLoader) processStruct(v reflect.Value) error {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Handle nested structs
		if field.Type.Kind() == reflect.Struct {
			if err := c.processStruct(fieldValue); err != nil {
				return err
			}
			continue
		}

		tag := field.Tag.Get("env")
		if tag == "" {
			continue
		}

		defaultValue := field.Tag.Get("default")
		value, err := c.resolveValue(tag, defaultValue)
		if err != nil {
			return err
		}

		if err := c.setField(fieldValue, value); err != nil {
			return fmt.Errorf("failed to set field %s: %w", field.Name, err)
		}
	}

	return nil
}

func (c *ConfigLoader) resolveValue(envName, defaultValue string) (string, error) {

	// First: check if the environment variable is set
	if value := os.Getenv(envName); value != "" {
		return value, nil
	}

	// Second: check if the environment variable is set via file with _FILE suffix (used for docker secrets)
	if filePath := os.Getenv(envName + "_FILE"); filePath != "" {
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
		}
		return string(fileContent), nil
	}

	// Third: Check .env file variables
	if value, ok := c.envVars[envName]; ok {
		return value, nil
	}

	return defaultValue, nil
}

func (c *ConfigLoader) setField(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var intValue int64
		if _, err := fmt.Sscanf(value, "%d", &intValue); err != nil {
			return err
		}
		field.SetInt(intValue)
	case reflect.Bool:
		var boolValue bool
		if _, err := fmt.Sscanf(value, "%t", &boolValue); err != nil {
			return err
		}
		field.SetBool(boolValue)
	case reflect.Float64:
		var floatValue float64
		if _, err := fmt.Sscanf(value, "%f", &floatValue); err != nil {
			return err
		}
		field.SetFloat(floatValue)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}
	return nil
}
