// Package config provides all the config management
package config

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/go-crypt/crypt"
	"github.com/go-crypt/crypt/algorithm"
	"github.com/go-crypt/crypt/algorithm/bcrypt"
	"github.com/tidwall/sjson"

	"github.com/fclairamb/ftpserver/config/confpar"
	"github.com/fclairamb/ftpserver/fs"
)

// ErrUnknownUser is returned when the provided user cannot be identified through our authentication mechanism
var ErrUnknownUser = errors.New("unknown user")

// hashedPasswordPrefixes lists the prefixes of the password hash formats we support (md5crypt, bcrypt variants,
// sha256crypt and sha512crypt). A password without any of these prefixes is considered to be plain-text.
var hashedPasswordPrefixes = []string{"$1$", "$2$", "$2a$", "$2b$", "$2x$", "$2y$", "$5$", "$6$"}

// bcryptCost is the cost used when hashing plain-text passwords
const bcryptCost = 10

// passwordDecoder is shared between all the authentication requests, it's safe for concurrent use
var passwordDecoder = sync.OnceValues(crypt.NewDecoderAll)

// Config provides the general server config
type Config struct {
	fileName string
	logger   *slog.Logger
	Content  *confpar.Content
}

// NewConfig creates a new config instance
func NewConfig(fileName string, logger *slog.Logger) (*Config, error) {
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
func FromContent(content *confpar.Content, fileName string, logger *slog.Logger) (*Config, error) {
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

// isHashedPassword tells if a password is stored in one of the supported hashed formats
func isHashedPassword(pass string) bool {
	for _, prefix := range hashedPasswordPrefixes {
		if strings.HasPrefix(pass, prefix) {
			return true
		}
	}

	return false
}

// newHasher creates the hasher used to store plain-text passwords
func newHasher() (algorithm.Hash, error) {
	return bcrypt.New(bcrypt.WithCost(bcryptCost))
}

// HashPassword returns the bcrypt hash of a plain-text password, in the format expected in the config file
func HashPassword(pass string) (string, error) {
	hasher, err := newHasher()
	if err != nil {
		return "", err
	}

	digest, err := hasher.Hash(pass)
	if err != nil {
		return "", err
	}

	return digest.Encode(), nil
}

// HashPlaintextPasswords replaces all the plain-text passwords of the config file by their bcrypt equivalent,
// and saves the config file.
func (c *Config) HashPlaintextPasswords() error {
	fileInfo, errStat := os.Stat(c.fileName)
	if errStat != nil {
		c.logger.Error("Cannot stat config file", "err", errStat)

		return errStat
	}

	content, errReadFile := os.ReadFile(c.fileName)
	if errReadFile != nil {
		c.logger.Error("Cannot read config file", "err", errReadFile)

		return errReadFile
	}

	var hasher algorithm.Hash

	save := false

	for i, a := range c.Content.Accesses {
		if a.Fs == "keycloak" || (a.User == "anonymous" && a.Pass == "*") || isHashedPassword(a.Pass) {
			continue
		}

		if hasher == nil {
			var err error
			if hasher, err = newHasher(); err != nil {
				return err
			}
		}

		digest, err := hasher.Hash(a.Pass)
		if err != nil {
			return err
		}

		modified, errJSONSet := sjson.Set(string(content), fmt.Sprintf("accesses.%d.pass", i), digest.Encode())
		if errJSONSet != nil {
			return fmt.Errorf("could not update the password of access %d: %w", i, errJSONSet)
		}

		a.Pass = digest.Encode()
		content = []byte(modified)
		save = true
	}

	if save {
		// We keep the original file permissions as the file usually contains secrets
		//nolint:gosec // the file name comes from the command line
		if errWriteFile := os.WriteFile(c.fileName, content, fileInfo.Mode().Perm()); errWriteFile != nil {
			c.logger.Error("Cannot write config file", "err", errWriteFile)

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

// checkPassword checks a password against its stored (plain-text or hashed) version
func checkPassword(stored, pass string) (bool, error) {
	if !isHashedPassword(stored) {
		return subtle.ConstantTimeCompare([]byte(stored), []byte(pass)) == 1, nil
	}

	decoder, err := passwordDecoder()
	if err != nil {
		return false, err
	}

	digest, err := decoder.Decode(stored)
	if err != nil {
		return false, err
	}

	return digest.MatchAdvanced(pass)
}

// GetAccess return a file system access given some credentials. Accesses are evaluated in their declaration order.
func (c *Config) GetAccess(user string, pass string) (*confpar.Access, error) {
	for _, a := range c.Content.Accesses {
		if a.Fs == "keycloak" {
			// The credentials are checked by keycloak when loading the file system. We work on a copy because
			// the access is shared between all the connections.
			access := *a
			access.User = user
			access.Pass = pass

			return &access, nil
		}

		if a.User != user {
			continue
		}

		if a.User == "anonymous" && a.Pass == "*" {
			return a, nil
		}

		ok, err := checkPassword(a.Pass, pass)
		if err != nil {
			return nil, err
		}

		if ok {
			return a, nil
		}
	}

	return nil, ErrUnknownUser
}
