package confpar

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
