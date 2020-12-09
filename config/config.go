// Package config provides all the config management
package config

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/fclairamb/ftpserverlib/log"

	"github.com/fclairamb/ftpserver/config/confpar"
	"github.com/fclairamb/ftpserver/fs"
)

// ErrUnknownUser is returned when the provided user cannot be identified through our authentication mechanism
var ErrUnknownUser = errors.New("unknown user")

// Config provides the general server config
type Config struct {
	fileName string
	logger   log.Logger
	Content  *confpar.Content
}

// NewConfig creates a new config instance
func NewConfig(fileName string, logger log.Logger) (*Config, error) {
	if fileName == "" {
		fileName = "ftpserver.json"
	}

	config := &Config{
		fileName: fileName,
		logger:   logger,
	}

	if err := config.Load(); err != nil {
		return nil, err
	}

	return config, nil
}

// Load the config
func (c *Config) Load() error {
	file, errOpen := os.Open(c.fileName)

	if errOpen != nil {
		return errOpen
	}

	defer func() {
		if errClose := file.Close(); errClose != nil {
			c.logger.Error("Cannot close config file", "err", errClose)
		}
	}()

	decoder := json.NewDecoder(file)

	// We parse and then copy to allow hot-reload in the future
	var content confpar.Content
	if errDecode := decoder.Decode(&content); errDecode != nil {
		c.logger.Error("Cannot decode file", "err", errDecode)
		return errDecode
	}

	c.Content = &content

	return c.Prepare()
}

// Prepare the config before using it
func (c *Config) Prepare() error {
	ct := c.Content
	if ct.ListenAddress == "" {
		ct.ListenAddress = "0.0.0.0:2121"
	}

	if publicHost := os.Getenv("PUBLIC_HOST"); publicHost != "" {
		ct.PublicHost = publicHost
	}

	return nil
}

// CheckAccesses checks all accesses
func (c *Config) CheckAccesses() error {
	for _, access := range c.Content.Accesses {
		_, errAccess := fs.LoadFs(access, c.logger)
		if errAccess != nil {
			c.logger.Error("Config: Invalid access !", "err", errAccess, "username", access.User, "fs", access.Fs)
			return errAccess
		}
	}

	return nil
}

// GetAccess return a file system access given some credentials
func (c *Config) GetAccess(user string, pass string) (*confpar.Access, error) {
	for _, a := range c.Content.Accesses {
		if a.User == user && (a.Pass == pass || (a.User == "anonymous" && a.Pass == "*")) {
			return a, nil
		}
	}

	return nil, ErrUnknownUser
}
