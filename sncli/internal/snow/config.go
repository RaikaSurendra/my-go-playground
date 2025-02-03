	package snow

	import (
	  "encoding/json"
	  "fmt"
	  "os"
	  "path/filepath"
	)

	type Config struct {
	  Instance string `json:"instance"`
	  Username string `json:"username"`
	  Password string `json:"password"`
	}

	func getConfigPath() (string, error) {
	  homeDir, err := os.UserHomeDir()
	  if err != nil {
	    return "", fmt.Errorf("failed to get home directory: %v", err)
	  }
	  return filepath.Join(homeDir, ".sncli", "config.json"), nil
	}

	func ReadConfig() (*Config, error) {
	  configPath, err := getConfigPath()
	  if err != nil {
	    return nil, err
	  }

	  data, err := os.ReadFile(configPath)
	  if err != nil {
	    if os.IsNotExist(err) {
	      return nil, fmt.Errorf("no config file found at %s - please run 'connect' command first", configPath)
	    }
	    return nil, fmt.Errorf("failed to read config file: %v", err)
	  }

	  var config Config
	  if err := json.Unmarshal(data, &config); err != nil {
	    return nil, fmt.Errorf("failed to parse config file: %v", err)
	  }

	  return &config, nil
	}

	func SaveConfig(config *Config) error {
	  configPath, err := getConfigPath()
	  if err != nil {
	    return err
	  }

	  // Create config directory if it doesn't exist
	  configDir := filepath.Dir(configPath)
	  if err := os.MkdirAll(configDir, 0700); err != nil {
	    return fmt.Errorf("failed to create config directory: %v", err)
	  }

	  data, err := json.MarshalIndent(config, "", "  ")
	  if err != nil {
	    return fmt.Errorf("failed to marshal config: %v", err) 
	  }

	  if err := os.WriteFile(configPath, data, 0600); err != nil {
	    return fmt.Errorf("failed to write config file: %v", err)
	  }

	  return nil
	}

