package config

import (
	"encoding/json"
	"errors"
	"github.com/fclairamb/ftpserver/config/confpar"
	"github.com/fclairamb/ftpserver/fs"
	"github.com/fclairamb/ftpserverlib/log"
	"os"
)

// Config provides the general server config
type Config struct {
	fileName string
	logger   log.Logger
	Content  *confpar.Content
}

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

func (c *Config) Load() error {
	file, errOpen := os.Open(c.fileName)
	if errOpen != nil {
		return errOpen
	}
	defer func() {
		if errClose := file.Close(); errClose != nil {
			c.logger.Error("Cannot close config file", errClose)
		}
	}()
	decoder := json.NewDecoder(file)

	// We parse and then copy to allow hot-reload in the future
	var content confpar.Content
	if errDecode := decoder.Decode(&content); errDecode != nil {
		c.logger.Error("Cannot decode file", errDecode)
		return errDecode
	}
	c.Content = &content
	return c.Prepare()
}

func (c *Config) Prepare() error {
	ct := c.Content
	if ct.ListenAddress == "" {
		ct.ListenAddress = "0.0.0.0:2121"
	}

	return nil
	// return c.CheckAccesses()
}

func (c *Config) CheckAccesses() error {
	for _, access := range c.Content.Accesses {
		_, errAccess := fs.LoadFs(&access)
		if errAccess != nil {
			c.logger.Error("Config: Invalid access !", errAccess, "username", access.User, "fs", access.Fs)
			return errAccess
		}
	}
	return nil
}

func (c *Config) GetAccess(user string, pass string) (*confpar.Access, error) {
	for _, a := range c.Content.Accesses {
		if a.User == user && a.Pass == pass {
			return &a, nil
		}
	}
	return nil, errors.New("unknown user")
}
