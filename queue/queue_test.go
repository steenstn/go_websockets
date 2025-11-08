package queue

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQueue_PushAndPop(t *testing.T) {
	queue := NewQueue(3)
	queue.Push("first")
	queue.Push("second")
	queue.Push("third")
	result := queue.Pop()
	assert.Equal(t, result, "first")
	result = queue.Pop()
	assert.Equal(t, result, "second")
	result = queue.Pop()
	assert.Equal(t, result, "third")
}

func TestQueue_PopEmpty(t *testing.T) {
	queue := NewQueue(3)
	result := queue.Pop()
	assert.Equal(t, result, "")
}

func TestQueue_PushAndPopMore(t *testing.T) {
	queue := NewQueue(3)
	queue.Push("first")
	queue.Push("second")

	result := queue.Pop()
	assert.Equal(t, result, "first")

	queue.Push("third")
	result = queue.Pop()
	assert.Equal(t, result, "second")
	result = queue.Pop()
	assert.Equal(t, result, "third")
}

func TestQueue_FullQueue(t *testing.T) {
	queue := NewQueue(3)
	queue.Push("first")
	queue.Push("second")
	queue.Push("third")
	queue.Push("Too much")
	result := queue.Pop()
	assert.Equal(t, result, "first")
	result = queue.Pop()
	assert.Equal(t, result, "second")
	result = queue.Pop()
	assert.Equal(t, result, "third")
	result = queue.Pop()
	assert.Equal(t, result, "")

	queue.Push("first")
	queue.Push("second")

	result = queue.Pop()
	assert.Equal(t, result, "first")

	queue.Push("third")

	result = queue.Pop()
	assert.Equal(t, result, "second")

	queue.Push("fourth")
	result = queue.Pop()
	assert.Equal(t, result, "third")
	result = queue.Pop()
	assert.Equal(t, result, "fourth")
}

func TestQueue_PeekHead(t *testing.T) {
	queue := NewQueue(3)
	queue.Push("first")
	assert.Equal(t, "first", queue.Peek())

	queue.Push("second")
	assert.Equal(t, "first", queue.Peek())

	queue.Push("third")
	assert.Equal(t, "first", queue.Peek())
}

func TestQueue_PeekTail(t *testing.T) {
	queue := NewQueue(3)
	queue.Push("first")
	assert.Equal(t, "first", queue.PeekTail())

	queue.Push("second")
	assert.Equal(t, "second", queue.PeekTail())

	queue.Push("third")
	assert.Equal(t, "third", queue.PeekTail())

	queue.Push("fourth")
	assert.Equal(t, "third", queue.PeekTail())
}

func TestQueue_PeekTailWithPopping(t *testing.T) {
	queue := NewQueue(3)
	queue.Push("A")
	queue.Push("B")
	queue.Push("C")
	queue.Pop()
	queue.Push("D")
	assert.Equal(t, "D", queue.PeekTail())
}
