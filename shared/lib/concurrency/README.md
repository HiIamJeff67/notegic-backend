# Concurrency Library

## Overview

`shared/lib/concurrency` provides a small generic worker-pool utility for running a slice of inputs in parallel while keeping result order stable.

This package is useful when you want to:

- fan out independent CPU or I/O tasks,
- cap worker count,
- and still map results back to the original input index.

## Files

- `concurrency.go`: generic data structures and processor function type.
- `execution.go`: worker-pool execution logic.
- `limit.go`: global worker count cap.

## Public API

```go
type Job[T any]
type Result[R any]
type ProcessorFunc[T any, R any] func(data T) (R, error)
func Execute[T any, R any](inputs []T, workerCount int, processorFunc ProcessorFunc[T, R]) []Result[R]
const MaxNumOfWorker = 1000
```

## Behavior Notes

- Returns an empty slice when `inputs` is empty.
- Actual worker count is clamped to `min(workerCount, len(inputs), MaxNumOfWorker)`.
- Result ordering is deterministic by input index, even though processing is concurrent.
- Any error returned by `processorFunc` is stored in `Result.Err` for that index.

## Example

```go
package main

import (
	"fmt"

	concurrency "github.com/HiIamJeff67/notegic-backend/shared/lib/concurrency"
)

func main() {
	inputs := []int{2, 4, 6, 8}

	results := concurrency.Execute(inputs, 3, func(v int) (int, error) {
		return v * v, nil
	})

	for _, r := range results {
		if r.Err != nil {
			fmt.Printf("index=%d error=%v\n", r.Index, r.Err)
			continue
		}
		fmt.Printf("index=%d input=%d output=%d\n", r.Index, inputs[r.Index], r.Data)
	}
}
```

## Project Usage Example

This package is used in service flows such as block validation fan-out in:

- `runtimes/core/services/block_service.go`

## File Structure

```text
shared/lib/concurrency/
├── README.md
├── LICENSE.md
├── concurrency.go
├── execution.go
└── limit.go
```
