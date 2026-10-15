package telegram

import (
	"context"
	"sync"
)

// queue is an unbounded FIFO of message IDs. push never blocks, so the
// updates loop that feeds it never waits on a forward in progress.
type queue struct {
	mu    sync.Mutex
	ids   []int
	ready chan struct{}
}

func newQueue() *queue {
	return &queue{ready: make(chan struct{}, 1)}
}

func (q *queue) push(id int) {
	q.mu.Lock()
	q.ids = append(q.ids, id)
	q.mu.Unlock()
	select {
	case q.ready <- struct{}{}:
	default:
	}
}

// front waits for the oldest ID without removing it, so an ID whose forward
// is interrupted by a restart is retried on the next connection.
func (q *queue) front(ctx context.Context) (int, error) {
	for {
		q.mu.Lock()
		if len(q.ids) > 0 {
			id := q.ids[0]
			q.mu.Unlock()
			return id, nil
		}
		q.mu.Unlock()
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-q.ready:
		}
	}
}

// pop removes the oldest ID.
func (q *queue) pop() {
	q.mu.Lock()
	q.ids = q.ids[1:]
	q.mu.Unlock()
}
