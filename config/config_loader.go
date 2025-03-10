package config

import (
	"fmt"
	"log/slog"
	"os"
	"reflect"

	"github.com/hashicorp/go-envparse"
	"gopkg.in/yaml.v3"
)

// EnvGetter is a function type for retrieving environment variables
type EnvGetter func(string) string

type Loader struct {
	filePath  string
	envGetter EnvGetter
}

type LoaderOpts struct {
	ConfigFilePath string
	EnvGetter      EnvGetter
}

// NewLoader creates a new configuration loader
func NewLoader(opts *LoaderOpts) *Loader {
	if opts == nil {
		opts = &LoaderOpts{}
	}

	// If no file path is provided, try to find default config files
	filePath := opts.ConfigFilePath
	if filePath == "" {
		filePath = findDefaultConfigFile()
	}

	// Use the provided env getter or default to os.Getenv
	envGetter := opts.EnvGetter
	if envGetter == nil {
		envGetter = os.Getenv
	}

	return &Loader{
		filePath:  filePath,
		envGetter: envGetter,
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

// Load loads configuration for the application.
// It will use the following order of precedence:
//  1. Environment variables
//  2. Docker secrets (using _FILE suffix)
//  3. application.yaml or application.yml file
//  4. .env file
//  5. Default values from struct tags
func (c *Loader) Load(cfg any) error {
	// First, set default values from struct tags
	if err := c.loadDefaults(reflect.ValueOf(cfg).Elem()); err != nil {
		return fmt.Errorf("failed to load defaults: %w", err)
	}

	// Then, try to load from file if available
	if c.filePath != "" {
		file, err := c.openFile()
		if err != nil {
			slog.Warn(fmt.Sprintf("failed to open file: %s", c.filePath), "error", err)
		} else {
			defer func(file *os.File) {
				err = file.Close()
				if err != nil {
					slog.Error(fmt.Sprintf("failed to close file: %s", c.filePath), "error", err)
				}
			}(file)

			// Try YAML first
			yamlOk, yamlErr := c.loadYamlFile(file, cfg)
			if !yamlOk {
				envOk, envErr := c.loadEnvFile(file)
				if envOk {
					if err = c.processStruct(reflect.ValueOf(cfg).Elem()); err != nil {
						return err
					}
				} else {
					slog.Info("failed to load file as YAML or ENV format",
						"yamlError", yamlErr,
						"envError", envErr)
				}
			}
		}
	}

	// Finally, override with environment variables if available
	return c.loadEnvironmentVars(reflect.ValueOf(cfg).Elem())
}

// openFile opens the configuration file
func (c *Loader) openFile() (*os.File, error) {
	file, err := os.Open(c.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	return file, nil
}

// loadDefaults sets default values from struct tags
func (c *Loader) loadDefaults(v reflect.Value) error {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Handle nested structs recursively
		if field.Type.Kind() == reflect.Struct {
			if err := c.loadDefaults(fieldValue); err != nil {
				return err
			}
			continue
		}

		// Get default value from tag
		tag := field.Tag.Get("env")
		if tag == "" {
			continue
		}

		defaultValue := field.Tag.Get("default")
		if defaultValue != "" {
			if err := c.setField(fieldValue, defaultValue); err != nil {
				return fmt.Errorf("failed to set default value for %s: %w", field.Name, err)
			}
		}
	}
	return nil
}

// loadYamlFile loads YAML directly into the struct
func (c *Loader) loadYamlFile(file *os.File, cfg any) (bool, error) {
	if _, err := file.Seek(0, 0); err != nil {
		return false, fmt.Errorf("failed to reset file position: %w", err)
	}

	data, err := os.ReadFile(c.filePath)
	if err != nil {
		return false, fmt.Errorf("failed to read YAML file: %w", err)
	}

	// Unmarshal directly into the struct
	if err = yaml.Unmarshal(data, cfg); err != nil {
		return false, fmt.Errorf("failed to parse YAML file: %w", err)
	}

	return true, nil
}

// loadEnvFile loads environment variables from a .env file
func (c *Loader) loadEnvFile(file *os.File) (bool, error) {
	if _, err := file.Seek(0, 0); err != nil {
		return false, fmt.Errorf("failed to reset file position: %w", err)
	}

	vars, err := envparse.Parse(file)
	if err != nil {
		return false, fmt.Errorf("failed to parse ENV file: %w", err)
	}

	// Set environment variables from the file
	for k, v := range vars {
		err = os.Setenv(k, v)
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

// loadEnvironmentVars overrides values with environment variables
func (c *Loader) loadEnvironmentVars(v reflect.Value) error {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Handle nested structs recursively
		if field.Type.Kind() == reflect.Struct {
			if err := c.loadEnvironmentVars(fieldValue); err != nil {
				return err
			}
			continue
		}

		tag := field.Tag.Get("env")
		if tag == "" {
			continue
		}

		// Check environment variables
		if value := c.envGetter(tag); value != "" {
			if err := c.setField(fieldValue, value); err != nil {
				return fmt.Errorf("failed to set field %s from env: %w", field.Name, err)
			}
			continue
		}

		// Check for _FILE suffix (used for Docker secrets)
		if filePath := c.envGetter(tag + "_FILE"); filePath != "" {
			fileContent, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", filePath, err)
			}
			if err := c.setField(fieldValue, string(fileContent)); err != nil {
				return fmt.Errorf("failed to set field %s from file: %w", field.Name, err)
			}
		}
	}
	return nil
}

// processStruct sets values from environment variables into the struct
func (c *Loader) processStruct(v reflect.Value) error {
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

		// Check for _FILE suffix (used for Docker secrets)
		if filePath := c.envGetter(tag + "_FILE"); filePath != "" {
			fileContent, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", filePath, err)
			}
			if err := c.setField(fieldValue, string(fileContent)); err != nil {
				return fmt.Errorf("failed to set field %s from file: %w", field.Name, err)
			}
		}

		// Look for environment variables
		if value := c.envGetter(tag); value != "" {
			if err := c.setField(fieldValue, value); err != nil {
				return fmt.Errorf("failed to set field %s: %w", field.Name, err)
			}
		}
	}

	return nil
}

// setField sets the value of a field based on its type
func (c *Loader) setField(field reflect.Value, value string) error {
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
