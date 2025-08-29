// Package confpar provide the core parameters of the config
package confpar

import "time"

// Access provides rules around any access
type Access struct {
	User          string            `json:"user"`            // User authenticating
	Pass          string            `json:"pass"`            // Password used for authentication
	Fs            string            `json:"fs"`              // Backend used for accessing file
	Params        map[string]string `json:"params"`          // Backend parameters
	Logging       Logging           `json:"logging"`         // Logging parameters
	ReadOnly      bool              `json:"read_only"`       // Read-only access
	Shared        bool              `json:"shared"`          // Shared FS instance
	SyncAndDelete *SyncAndDelete    `json:"sync_and_delete"` // Local empty directory and synchronization
}

// AccessesWebhook defines an optional webhook to get user's access
type AccessesWebhook struct {
	URL     string            `json:"url"`     // URL to call
	Headers map[string]string `json:"headers"` // Token to use in the
	Timeout time.Duration     `json:"timeout"` // Max time request can take
}

// SyncAndDelete provides
type SyncAndDelete struct {
	Enable    bool   `json:"enable"`    // Instant write
	Directory string `json:"directory"` // Directory
}

// PortRange defines a port-range
// ... used only for the passive transfer listening range at this stage.
type PortRange struct {
	Start int `json:"start"` // Start of the range
	End   int `json:"end"`   // End of the range
}

// Logging defines how we will log accesses
type Logging struct {
	FtpExchanges bool   `json:"ftp_exchanges"` // Log all ftp exchanges
	FileAccesses bool   `json:"file_accesses"` // Log all file accesses
	File         string `json:"file"`          // Log file
}

// TLS define the TLS Config
type TLS struct {
	ServerCert *ServerCert `json:"server_cert"` // Server certificates
}

// ServerCert defines the TLS server certificate config
type ServerCert struct {
	Cert string `json:"cert"` // Public certificate(s)
	Key  string `json:"key"`  // Private key
}

// Extensions define the relevant configurations for extended features
type Extensions struct {
	EnableHASH bool `json:"enable_hash"` // Enable support for calculating hash value of files
}

// Content defines the content of the config file
type Content struct {
	Version                  int              `json:"version"`                     // File format version
	ListenAddress            string           `json:"listen_address"`              // Address to listen on
	PublicHost               string           `json:"public_host"`                 // Public host to listen on
	MaxClients               int              `json:"max_clients"`                 // Maximum clients who can connect
	HashPlaintextPasswords   bool             `json:"hash_plaintext_passwords"`    // Overwrite plain-text passwords with hashed equivalents
	Accesses                 []*Access        `json:"accesses"`                    // Accesses offered to users
	PassiveTransferPortRange *PortRange       `json:"passive_transfer_port_range"` // Listen port range
	Extensions               Extensions       `json:"extensions"`                  // Extended features
	Logging                  Logging          `json:"logging"`                     // Logging parameters
	TLS                      *TLS             `json:"tls"`                         // TLS Config
	TLSRequired              string           `json:"tls_required"`
	AccessesWebhook          *AccessesWebhook `json:"accesses_webhook"` // Webhook to call when accesses are updated
}
