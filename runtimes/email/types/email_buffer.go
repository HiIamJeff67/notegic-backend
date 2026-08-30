package types

import "container/heap"

/* ============================== Priority Queue for the Buffer of Email Task Manager ============================== */
// the below implementations are for the heap interface(heap.Interface)
// so the EmailBuffer agree with the heap.Interface, and it can be passed into heap.Push() and heap.Pop()

type EmailBuffer []*EmailTask

func (buffer EmailBuffer) Len() int {
	return len(buffer)
}

func (buffer EmailBuffer) Less(i, j int) bool {
	if buffer[i].Priority != buffer[j].Priority {
		return buffer[i].Priority > buffer[j].Priority
	}
	return buffer[i].CreatedAt.Before(buffer[j].CreatedAt)
}

func (buffer EmailBuffer) Swap(i, j int) {
	buffer[i], buffer[j] = buffer[j], buffer[i]
}

func (buffer *EmailBuffer) Push(x interface{}) {
	task := x.(*EmailTask)
	*buffer = append(*buffer, task)
}

func (buffer *EmailBuffer) Pop() interface{} {
	old := *buffer
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*buffer = old[0 : n-1]
	return item
}

/* ============================== Customized Methods for the EmailBuffer ============================== */

func (buffer *EmailBuffer) EnqueueTask(task *EmailTask) {
	heap.Push(buffer, task)
}

func (buffer *EmailBuffer) DequeueTask() *EmailTask {
	if buffer.Len() == 0 {
		return nil
	}

	return heap.Pop(buffer).(*EmailTask)
}

func (buffer EmailBuffer) Size() int {
	return len(buffer)
}

func (buffer EmailBuffer) IsEmpty() bool {
	return len(buffer) == 0
}

func (buffer EmailBuffer) Top() *EmailTask {
	if len(buffer) == 0 {
		return nil
	}
	return buffer[0]
}

func NewEmailBuffer() *EmailBuffer {
	buffer := &EmailBuffer{}
	heap.Init(buffer)
	return buffer
}
