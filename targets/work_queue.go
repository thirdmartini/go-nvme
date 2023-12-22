package targets

import (
	"fmt"
)

type Handler interface {
	Handle(r *IORequest)
}

type WorkQueue struct {
	queue   chan *IORequest
	Handler Target
}

func (w *WorkQueue) Close() error {
	close(w.queue)
	for r := range w.queue {
		fmt.Printf("Draining queue\n")
		r.Complete(TargetErrorAborted)
	}
	return w.Handler.Close()
}

func (w *WorkQueue) Work() {
	for {
		// Implement clean exit
		r := <-w.queue
		if r == nil {
			return
		}
		w.Handler.Queue(r)
	}
}

func (w *WorkQueue) Start() error {
	for i := 0; i < cap(w.queue); i++ {
		go w.Work()
	}
	return nil
}

func (w *WorkQueue) Queue(r *IORequest) TargetError {
	w.queue <- r
	return TargetErrorNone
}

// GetSize in bytes
func (w *WorkQueue) GetSize() uint64 {
	return w.Handler.GetSize()
}

func (w *WorkQueue) GetRuntimeDetails() []KV {
	return w.Handler.GetRuntimeDetails()
}

func NewWorkQueue(options map[string]string, h Target) *WorkQueue {
	return &WorkQueue{
		Handler: h,
		queue:   make(chan *IORequest, 8),
	}
}
