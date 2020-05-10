package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fclairamb/ftpserverlib/log"
	"github.com/spf13/afero"
	"os"
)

// Access provides rules around any access
type Access struct {
	User   string            `json:"user"`   // User authenticating
	Pass   string            `json:"pass"`   // Password used for authentication
	Fs     string            `json:"fs"`     // Backend used for accessing file
	Params map[string]string `json:"params"` // Backend parameters
}

// Content defines the content of the config file
type Content struct {
	Version       int      `json:"version"`        // File format version
	ListenAddress string   `json:"listen_address"` // Address to listen on
	MaxClients    int      `json:"max_clients"`    // Maximum clients who can connect at any given time
	Accesses      []Access `json:"accesses"`       // Accesses offered to users
}

// Config provides the general server config
type Config struct {
	fileName string
	logger   log.Logger
	Content  *Content
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
	var content Content
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
}

func (c *Config) GetAccess(user string, pass string) (*Access, error) {
	for _, a := range c.Content.Accesses {
		if a.User == user && a.Pass == pass {
			return &a, nil
		}
	}
	return nil, errors.New("unknown user")
}

func (a *Access) Check() error {
	_, err := a.GetFs()
	return err
}

func (a *Access) GetFs() (afero.Fs, error) {
	if a.Fs == "os" {
		basePath := a.Params["basePath"]
		if basePath == "" {
			return nil, errors.New("basePath must be specified")
		}
		return afero.NewBasePathFs(afero.NewOsFs(), basePath), nil
	}
	return nil, fmt.Errorf("unknown fs: %s", a.Fs)
}
