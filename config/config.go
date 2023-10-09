// Package config provides all the config management
package config

import (
	"encoding/json"
	"errors"
	"os"

	log "github.com/fclairamb/go-log"

	"github.com/clicknclear/ftpserver/config/confpar"
	"github.com/clicknclear/ftpserver/fs"
)

// ErrUnknownUser is returned when the provided user cannot be identified through our authentication mechanism
var ErrUnknownUser = errors.New("unknown user")

// Config provides the general server config
type Config struct {
	fileName  string
	logger    log.Logger
	Content   *confpar.Content
	accessMap map[string]*confpar.Access
}

// NewConfig creates a new config instance
func NewConfig(fileName string, logger log.Logger) (*Config, error) {
	if fileName == "" {
		fileName = "ftpserver.json"
	}

	config := &Config{
		fileName:  fileName,
		logger:    logger,
		accessMap: make(map[string]*confpar.Access),
	}

	if err := config.Load(); err != nil {
		return nil, err
	}

	return config, nil
}

// FromContent creates a new config instance from a pre-created Content and logger. The
// fileName should indicate origin of the given Content, but the file will never be opened.
func FromContent(content *confpar.Content, fileName string, logger log.Logger) (*Config, error) {
	c := &Config{
		fileName:  fileName,
		logger:    logger,
		Content:   content,
		accessMap: make(map[string]*confpar.Access),
	}

	if err := c.Prepare(); err != nil {
		return nil, err
	}

	return c, nil
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

	for _, access := range c.Content.Accesses {
		c.UpsertAccess(access)
	}

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
	access := c.accessMap[user]
	if access == nil {
		return nil, ErrUnknownUser
	}

	if access.Pass == pass || (access.User == "anonymous" && access.Pass == "*") {
		return access, nil
	}

	return nil, ErrUnknownUser
}

// GetAccess return a file system access given some credentials
func (c *Config) UpsertAccess(newAccess *confpar.Access) error {
	_, errAccess := fs.LoadFs(newAccess, c.logger)
	if errAccess != nil {
		c.logger.Error("Config: Invalid access !", "err", errAccess, "username", newAccess.User, "fs", newAccess.Fs)
		return errAccess
	}
	c.accessMap[newAccess.User] = newAccess
	return nil
}
