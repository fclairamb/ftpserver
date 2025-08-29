// Package config provides all the config management
package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	log "github.com/fclairamb/go-log"

	"github.com/fclairamb/ftpserver/config/confpar"
	"github.com/fclairamb/ftpserver/fs"

	"github.com/go-crypt/crypt"
	"github.com/go-crypt/crypt/algorithm/bcrypt"
	"github.com/tidwall/sjson"
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

// FromContent creates a new config instance from a pre-created Content and logger. The
// fileName should indicate origin of the given Content, but the file will never be opened.
func FromContent(content *confpar.Content, fileName string, logger log.Logger) (*Config, error) {
	c := &Config{
		fileName: fileName,
		logger:   logger,
		Content:  content,
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

	if c.Content.HashPlaintextPasswords {
		if err := c.HashPlaintextPasswords(); err != nil {
			return err
		}
	}

	return c.Prepare()
}

func (c *Config) HashPlaintextPasswords() error {
	json, errReadFile := os.ReadFile(c.fileName)
	if errReadFile != nil {
		c.logger.Error("Cannot read config file!", "err", errReadFile)
		return errReadFile
	}

	save := false
	for i, a := range c.Content.Accesses {
		if a.User == "anonymous" && a.Pass == "*" {
			continue
		}

		switch true {
		case bytes.HasPrefix([]byte(a.Pass), []byte("$1$")):
			//This user's password is md5crypt
			continue
		case bytes.HasPrefix([]byte(a.Pass), []byte("$2$")):
			//This user's password is bcrypt
			continue
		case bytes.HasPrefix([]byte(a.Pass), []byte("$2a$")):
			//This user's password is bcrypt-a
			continue
		case bytes.HasPrefix([]byte(a.Pass), []byte("$2b$")):
			//This user's password is bcrypt-b
			continue
		case bytes.HasPrefix([]byte(a.Pass), []byte("$2x$")):
			//This user's password is bcrypt-x
			continue
		case bytes.HasPrefix([]byte(a.Pass), []byte("$2y$")):
			//This user's password is bcrypt-y
			continue
		case bytes.HasPrefix([]byte(a.Pass), []byte("$5$")):
			//This user's password is sha256crypt
			continue
		case bytes.HasPrefix([]byte(a.Pass), []byte("$6$")):
			//This user's password is sha512crypt
			continue
		default:
			//This password is not hashed
			hasher, err := bcrypt.New(bcrypt.WithCost(10))
			if err != nil {
				return err
			}

			digest, err := hasher.Hash(a.Pass)
			if err != nil {
				return err
			}

			modified, errJsonSet := sjson.Set(string(json), "accesses."+fmt.Sprint(i)+".pass", digest.Encode())
			c.Content.Accesses[i].Pass = digest.Encode()
			if errJsonSet == nil {
				save = true
				json = []byte(modified)
			}
		}
	}
	if save {
		errWriteFile := os.WriteFile(c.fileName, json, 0644)
		if errWriteFile != nil {
			c.logger.Error("Cannot write config file!", "err", errWriteFile)
			return errWriteFile
		}
	}
	return nil
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
	decoder, err := crypt.NewDecoderAll()
	if err != nil {
		return nil, err
	}

	for _, a := range c.Content.Accesses {
		if a.Fs == "keycloak" {
			a.User = user
			a.Pass = pass
			return a, nil
		}

		if a.User == user {
			switch true {
			case bytes.HasPrefix([]byte(a.Pass), []byte("$1$")):
				//This user's password is md5crypt
				digest, err := decoder.Decode(a.Pass)
				if err != nil {
					return nil, err
				}

				ok, err := digest.MatchAdvanced(pass)
				if err != nil {
					return nil, err
				}

				if ok {
					return a, nil
				}
			case bytes.HasPrefix([]byte(a.Pass), []byte("$2$")):
				//This user's password is bcrypt
				digest, err := decoder.Decode(a.Pass)
				if err != nil {
					return nil, err
				}

				ok, err := digest.MatchAdvanced(pass)
				if err != nil {
					return nil, err
				}

				if ok {
					return a, nil
				}
			case bytes.HasPrefix([]byte(a.Pass), []byte("$2a$")):
				//This user's password is bcrypt-a
				digest, err := decoder.Decode(a.Pass)
				if err != nil {
					return nil, err
				}

				ok, err := digest.MatchAdvanced(pass)
				if err != nil {
					return nil, err
				}

				if ok {
					return a, nil
				}
			case bytes.HasPrefix([]byte(a.Pass), []byte("$2b$")):
				//This user's password is bcrypt-b
				digest, err := decoder.Decode(a.Pass)
				if err != nil {
					return nil, err
				}

				ok, err := digest.MatchAdvanced(pass)
				if err != nil {
					return nil, err
				}

				if ok {
					return a, nil
				}
			case bytes.HasPrefix([]byte(a.Pass), []byte("$2x$")):
				//This user's password is bcrypt-x
				digest, err := decoder.Decode(a.Pass)
				if err != nil {
					return nil, err
				}

				ok, err := digest.MatchAdvanced(pass)
				if err != nil {
					return nil, err
				}

				if ok {
					return a, nil
				}
			case bytes.HasPrefix([]byte(a.Pass), []byte("$2y$")):
				//This user's password is bcrypt-y
				digest, err := decoder.Decode(a.Pass)
				if err != nil {
					return nil, err
				}

				ok, err := digest.MatchAdvanced(pass)
				if err != nil {
					return nil, err
				}

				if ok {
					return a, nil
				}
			case bytes.HasPrefix([]byte(a.Pass), []byte("$5$")):
				//This user's password is sha256crypt
				digest, err := decoder.Decode(a.Pass)
				if err != nil {
					return nil, err
				}

				ok, err := digest.MatchAdvanced(pass)
				if err != nil {
					return nil, err
				}

				if ok {
					return a, nil
				}
			case bytes.HasPrefix([]byte(a.Pass), []byte("$6$")):
				//This user's password is sha512crypt
				digest, err := decoder.Decode(a.Pass)
				if err != nil {
					return nil, err
				}

				ok, err := digest.MatchAdvanced(pass)
				if err != nil {
					return nil, err
				}

				if ok {
					return a, nil
				}
			default:
				//This user's password is plain-text
				if a.Pass == pass || (a.User == "anonymous" && a.Pass == "*") {
					return a, nil
				}
			}
		}

	}

	return nil, ErrUnknownUser
}
