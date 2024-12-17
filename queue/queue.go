package queue

import "sync"

// TODO make generic

/*
Queue that uses a "circular array"
*/
type Queue struct {
	entries    []string
	head       int
	tail       int
	numEntries int
}

var lock sync.Mutex

func (q *Queue) Push(value string) {
	lock.Lock()
	defer lock.Unlock()
	if q.numEntries >= len(q.entries) {
		return
	}
	q.entries[q.tail] = value
	q.tail = (q.tail + 1) % (len(q.entries))

	q.numEntries++

	println()
}

func (q *Queue) Pop() string {
	lock.Lock()
	defer lock.Unlock()
	if q.numEntries <= 0 {
		return ""
	}
	value := q.entries[q.head]
	q.head = (q.head + 1) % (len(q.entries))
	q.numEntries--
	return value
}

func (q *Queue) Peek() string {
	if q.numEntries <= 0 {
		return ""
	}
	return q.entries[q.head]
}

func NewQueue(size int) Queue {
	return Queue{make([]string, size), 0, 0, 0}
}
