package resource

import (
	"strings"
	"sync"
)

var resourceMu sync.RWMutex

var resources = make(map[string]Factory)

func RegisterResource(name string, resource Factory) {
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

func Register(name string, resource Factory) {
	RegisterResource(name, resource)
}

func GetResource(name string) (Factory, bool) {
	name = strings.ToLower(name)
	resourceMu.RLock()
	defer resourceMu.RUnlock()
	resource, ok := resources[name]
	return resource, ok
}
