// Package confpar provide the core parameters of the config
package confpar

// Access provides rules around any access
type Access struct {
	User     string            `json:"user"`      // User authenticating
	Pass     string            `json:"pass"`      // Password used for authentication
	Fs       string            `json:"fs"`        // Backend used for accessing file
	Params   map[string]string `json:"params"`    // Backend parameters
	Logging  Logging           `json:"logging"`   // Logging parameters
	ReadOnly bool              `json:"read_only"` // Read-only access
}

// PortRange defines a port-range
// ... used only for the passive transfer listening range at this stage.
type PortRange struct {
	Start int `json:"start"` // Start of the range
	End   int `json:"end"`   // End of the range
}

// Logging defines how we will log accesses
type Logging struct {
	FtpExchanges bool `json:"ftp_exchanges"` // Log all ftp exchanges
	FileAccesses bool `json:"file_accesses"` // Log all file accesses
}

// Content defines the content of the config file
type Content struct {
	Version                  int        `json:"version"`                     // File format version
	ListenAddress            string     `json:"listen_address"`              // Address to listen on
	PublicHost               string     `json:"public_host"`                 // Public host to listen on
	MaxClients               int        `json:"max_clients"`                 // Maximum clients who can connect
	Accesses                 []*Access  `json:"accesses"`                    // Accesses offered to users
	PassiveTransferPortRange *PortRange `json:"passive_transfer_port_range"` // Listen port range
	Logging                  Logging    `json:"logging"`                     // Logging parameters
}
