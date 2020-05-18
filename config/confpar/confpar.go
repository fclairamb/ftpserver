// Package confpar provide the core parameters of the config
package confpar

// Access provides rules around any access
type Access struct {
	User   string            `json:"user"`   // User authenticating
	Pass   string            `json:"pass"`   // Password used for authentication
	Fs     string            `json:"fs"`     // Backend used for accessing file
	Params map[string]string `json:"params"` // Backend parameters
}

// Listen port range
type PortRange struct {
	Start int `json:"start"` // Start of the range
	End   int `json:"end"`   // End of the range
}

// Content defines the content of the config file
type Content struct {
	Version                  int        `json:"version"`                     // File format version
	ListenAddress            string     `json:"listen_address"`              // Address to listen on
	MaxClients               int        `json:"max_clients"`                 // Maximum clients who can connect at any given time
	Accesses                 []*Access  `json:"accesses"`                    // Accesses offered to users
	PassiveTransferPortRange *PortRange `json:"passive_transfer_port_range"` // Listen port range
}
