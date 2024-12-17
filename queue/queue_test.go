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
