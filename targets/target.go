package targets

// KV contains a key value pair
type KV struct {
	Key   string
	Value string
}

type Target interface {
	// Start performs any required initialization of the target
	Start() error
	// Queue enqueue the request for processing
	Queue(r *IORequest) TargetError
	// GetSize the current size (in bytes) of the target
	GetSize() uint64
	// Close terminates any background tasks the target m ay be doing
	Close() error
	// GetRuntimeDetails returns internal configuration information about the target
	GetRuntimeDetails() []KV
}
