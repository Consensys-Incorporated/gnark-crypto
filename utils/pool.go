// Copyright 2020-2026 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package utils

import (
	"fmt"
	"math/bits"
	"runtime"
	"sync"
)

// Pool is a pool of slices of T, grouped by power-of-two capacity. Its copies share the same slices.
// Dumped slices are retained for the lifetime of the Pool and are not evicted by the garbage collector,
// so a Pool should be scoped to an operation rather than kept globally.
type Pool[T any] struct {
	*pool[T]
}

type pool[T any] struct {
	lock     sync.Mutex
	subPools          [bits.UintSize][][]T
	trackingUndumpted bool
	undumped          map[*T][]uintptr // call stack of Make for each tracked slice not yet dumped, keyed by data pointer
}

// StartTrackingUndumped starts recording the call site of every slice handed out by Make, until the slice is
// dumped, so that PrintUndumped can report slices that were never dumped. It is a development aid for finding
// missing calls to Dump.
func (p *pool[T]) StartTrackingUndumped() {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.trackingUndumpted = true
}

// PauseTrackingUndumped stops recording slices handed out by Make. Slices already recorded are still removed
// when dumped.
func (p *pool[T]) PauseTrackingUndumped() {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.trackingUndumpted = false
}

// NewPool returns a pool with a size class for every power of two. maxN is ignored; it is kept for compatibility.
func NewPool[T any](maxN ...int) Pool[T] {
	_ = maxN
	p := new(pool[T])
	p.undumped = make(map[*T][]uintptr)
	return Pool[T]{p}
}

// Make a new slice of size n, reusing a free existing one if possible.
func (p *pool[T]) Make(n int) []T {
	if n <= 0 {
		return nil
	}

	poolI := ceilLog2(n)

	p.lock.Lock()
	if stack := p.subPools[poolI]; len(stack) > 0 {
		last := len(stack) - 1
		slice := stack[last]
		p.subPools[poolI] = stack[:last]
		if p.trackingUndumpted {
			p.trackMake(slice)
		}
		p.lock.Unlock()
		return slice[:n]
	}

	trackingUndumped := p.trackingUndumpted
	p.lock.Unlock()

	slice := make([]T, 0, 1<<poolI)

	if trackingUndumped {
		p.lock.Lock()
		p.trackMake(slice)
		p.lock.Unlock()
	}

	return slice[:n]
}

// trackMake records the call site of Make for a slice it handed out. The caller must hold the lock.
func (p *pool[T]) trackMake(slice []T) {
	pcs := make([]uintptr, 2)
	p.undumped[getDataPointer(slice)] = pcs[:runtime.Callers(3, pcs)]
}

// Dump a set of slices into the pool. It doesn't need to have been created by the pool, but it must similarly
// have a power of 2 as its capacity.
func (p *pool[T]) Dump(slices ...[]T) {
	p.lock.Lock()
	defer p.lock.Unlock()

	for _, slice := range slices {
		if cap(slice) == 0 {
			continue
		}
		ptr := getDataPointer(slice)
		poolI := ceilLog2(cap(slice))

		if cap(slice) != 1<<poolI {
			panic("pool slice capacity must be a power of 2")
		}

		for _, free := range p.subPools[poolI] {
			if getDataPointer(free) == ptr {
				panic("slice double-dumped")
			}
		}
		p.subPools[poolI] = append(p.subPools[poolI], slice[:0])

		delete(p.undumped, ptr)
	}
}

func printFrame(frame runtime.Frame) {
	fmt.Printf("\t%s line %d, function %s\n", frame.File, frame.Line, frame.Function)
}

// ceilLog2 returns the smallest k such that 1<<k >= n. It requires n >= 1.
func ceilLog2(n int) int {
	return bits.Len(uint(n - 1))
}

// getDataPointer returns a pointer to the first element of slice's underlying array. It requires cap(slice) ≥ 1.
func getDataPointer[T any](slice []T) *T {
	return &slice[:1][0]
}

// PrintUndumped prints the call sites of Make for the slices handed out while trackingUndumpted that have not been dumped.
func (p *pool[T]) PrintUndumped() {
	p.lock.Lock()
	defer p.lock.Unlock()

	fmt.Printf("slices never dumped: %d\n", len(p.undumped))
	for _, pcs := range p.undumped {
		fmt.Println("-------------------------")

		var frame runtime.Frame
		frames := runtime.CallersFrames(pcs)
		more := true
		for more {
			frame, more = frames.Next()
			printFrame(frame)
		}
	}
}

func (p *pool[T]) Clone(slice []T) []T {
	res := p.Make(len(slice))
	copy(res, slice)
	return res
}
