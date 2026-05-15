package resource

import (
	"strings"
	"sync"
)

var resourceMu sync.RWMutex

var resources = make(map[string]Factory)

// Register registers a Factory for the given scheme name. It panics if the
// factory is nil or if a factory has already been registered for the name.
func Register(name string, resource Factory) {
	if resource == nil {
		panic("gonfig: RegisterResource resource is nil")
	}
	name = strings.ToLower(name)
	resourceMu.Lock()
	defer resourceMu.Unlock()
	if _, dup := resources[name]; dup {
		panic("gonfig: RegisterResource called twice for resource " + name)
	}
	resources[name] = resource
}

// Get returns the Factory registered for the given scheme name and a boolean
// indicating whether a factory was found.
func Get(name string) (Factory, bool) {
	name = strings.ToLower(name)
	resourceMu.RLock()
	defer resourceMu.RUnlock()
	resource, ok := resources[name]
	return resource, ok
}
