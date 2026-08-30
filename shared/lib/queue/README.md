# Queue Library

## Overview

`shared/lib/queue` implements a generic FIFO queue with fixed capacity.

It is a lightweight data structure used by Notegic internals (for example, BFS-like
traversals in adapters) and supports any element type through Go generics.

## Files

- `queue.go`: queue node type, queue struct, and queue operations.

## Public API

- `func NewQueue[T any](capacity int) Queue[T]`
- `func (q *Queue[T]) Enqueue(element T) error`
- `func (q *Queue[T]) Dequeue() (T, error)`
- `func (q *Queue[T]) Front() (T, error)`
- `func (q *Queue[T]) Size() int`
- `func (q *Queue[T]) Capacity() int`
- `func (q *Queue[T]) SetCapacity(capacity int) error`
- `func (q *Queue[T]) IsFull() bool`
- `func (q *Queue[T]) IsEmpty() bool`

## Behavior Notes

- Capacity is enforced by `Enqueue`; exceeding capacity returns `queue is full`.
- `Dequeue` and `Top` return `queue is empty` when there is no element.
- `SetCapacity` rejects values smaller than current queue size to avoid implicit data loss.
- The implementation is not thread-safe; synchronize externally for concurrent access.

## Example

```go
package main

import (
	"fmt"

	queue "github.com/HiIamJeff67/notegic-backend/shared/lib/queue"
)

func main() {
	q := queue.NewQueue[string](3)

	_ = q.Enqueue("A")
	_ = q.Enqueue("B")

	front, _ := q.Top()
	fmt.Println("front:", front) // A

	item, _ := q.Dequeue()
	fmt.Println("dequeue:", item) // A

	fmt.Println("size:", q.Size()) // 1
}
```

## Project Usage Example

This package is used in block tree flatten/arborize flows in:

- `runtimes/adapters/editable_block_adapter.go`

## File Structure

```text
shared/lib/queue/
├── README.md
├── LICENSE.md
└── queue.go
```
