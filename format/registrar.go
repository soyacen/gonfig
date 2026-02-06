package format

import (
	"strings"
	"sync"
)

// Global formatters registry mapping file extensions to their corresponding parsers
var (
	// formatters stores registered format parsers
	formatters = make(map[string]Formatter)
	// mutex to protect concurrent access to formatters
	mutex sync.RWMutex
)

// RegisterFormatter associates a file extension with a configuration parser
//
// Args:
//
//	ext (string): File extension (e.g., "yaml", "toml")
//	formatter (Formatter): Implementation of the Formatter interface
func RegisterFormatter(ext string, formatter Formatter) {
	if formatter == nil {
		panic("gonfig: RegisterFormatter formatter is nil")
	}
	ext = strings.ToLower(ext)
	mutex.Lock()
	defer mutex.Unlock()
	if _, dup := formatters[ext]; dup {
		panic("gonfig: RegisterFormatter called twice for extension " + ext)
	}
	formatters[ext] = formatter
}

// GetFormatter retrieves the parser associated with a specific file extension
//
// Args:
//
//	ext (string): File extension to look up
//
// Returns:
//
//	Formatter: Registered parser or nil if not found
func GetFormatter(ext string) (Formatter, bool) {
	ext = strings.ToLower(ext)
	mutex.RLock()
	defer mutex.RUnlock()
	formatter, ok := formatters[ext]
	return formatter, ok
}
