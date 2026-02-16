package scheduler

import (
	"container/heap"
	"sync"
	"time"
)

// TaskQueue implements heap.Interface and holds ReconciliationTasks.
type TaskQueue []*ReconciliationTask

func (pq TaskQueue) Len() int { return len(pq) }

func (pq TaskQueue) Less(i, j int) bool {
	// Anti-Starvation: Calculate Effective Priority
	// EffectivePriority = BasePriority - (WaitTime / AgingFactor)
	// We want Pop to give us the lowest effective priority value (highest urgency)

	now := time.Now()
	// AgingFactor: Every 10 seconds of waiting reduces priority value by 1 (improving precedence)
	const agingFactorSeconds = 10.0

	effPriI := float64(pq[i].Priority) - (now.Sub(pq[i].SubmitTime).Seconds() / agingFactorSeconds)
	effPriJ := float64(pq[j].Priority) - (now.Sub(pq[j].SubmitTime).Seconds() / agingFactorSeconds)

	// If effective priorities are roughly equal, use Deadline
	if int(effPriI) == int(effPriJ) {
		return pq[i].Deadline.Before(pq[j].Deadline)
	}
	return effPriI < effPriJ
}

func (pq TaskQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *TaskQueue) Push(x interface{}) {
	item := x.(*ReconciliationTask)
	*pq = append(*pq, item)
}

func (pq *TaskQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}

// ThreadSafeQueue wraps TaskQueue with a mutex for safe concurrent access.
type ThreadSafeQueue struct {
	pq TaskQueue
	mu sync.Mutex
}

func NewThreadSafeQueue() *ThreadSafeQueue {
	return &ThreadSafeQueue{
		pq: make(TaskQueue, 0),
	}
}

func (q *ThreadSafeQueue) Push(task *ReconciliationTask) {
	q.mu.Lock()
	defer q.mu.Unlock()
	heap.Push(&q.pq, task)
}

func (q *ThreadSafeQueue) Pop() *ReconciliationTask {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.pq) == 0 {
		return nil
	}
	return heap.Pop(&q.pq).(*ReconciliationTask)
}

func (q *ThreadSafeQueue) Peek() *ReconciliationTask {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.pq) == 0 {
		return nil
	}
	// Heap root is at index 0. Accessing safely.
	return q.pq[0]
}

func (q *ThreadSafeQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.pq)
}

// PushDelayed pushes a task to the queue after a delay.
// This is non-blocking.
func (q *ThreadSafeQueue) PushDelayed(task *ReconciliationTask, delay time.Duration) {
	time.AfterFunc(delay, func() {
		q.Push(task)
	})
}
