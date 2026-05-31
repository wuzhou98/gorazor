package razorcore

import "sync"

// LayoutCache provides thread-safe isolation for layout compilations
type LayoutCache struct {
	mutex sync.RWMutex
	m     map[string][]string
}

// NewLayoutCache instantiates a new isolated LayoutCache
func NewLayoutCache() *LayoutCache {
	return &LayoutCache{
		m: make(map[string][]string),
	}
}

// Get returns arguments of a given layout file from the cache
func (c *LayoutCache) Get(file string) []string {
	if c == nil {
		return []string{}
	}
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if args, ok := c.m[file]; ok {
		return args
	}
	return []string{}
}

// Set registers arguments for a layout file in the cache
func (c *LayoutCache) Set(file string, args []string) {
	if c == nil {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.m[file] = args
}
