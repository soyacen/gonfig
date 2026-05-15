package format

import (
	"strings"
	"sync"
)

var (
	formatters = make(map[string]Formatter)
	mutex      sync.RWMutex
)

// RegisterFormatter registers a Formatter for the given file extension. It
// panics if formatter is nil or if an extension is registered twice.
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

// GetFormatter returns the Formatter registered for the given file extension
// and a boolean indicating whether a formatter was found.
func GetFormatter(ext string) (Formatter, bool) {
	ext = strings.ToLower(ext)
	mutex.RLock()
	defer mutex.RUnlock()
	formatter, ok := formatters[ext]
	return formatter, ok
}
