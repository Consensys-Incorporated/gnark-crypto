// Copyright 2020-2026 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package utils

import (
	"encoding/json"
	"fmt"
	"math/bits"
	"runtime"
	"sync"
	"unsafe"
)

type sizedPoolStats struct {
	Used          int
	Allocated     int
	ReuseRate     float64
	InUse         int
	GreatestNUsed int
	SmallestNUsed int
}

type poolStats[T any] struct {
	recording    bool
	subPools     [64]sizedPoolStats
	allocatedFor map[*T][]uintptr
}

// Pool is a pool of slices of T, grouped by power-of-two capacity. Its copies share the same slices.
type Pool[T any] struct {
	*pool[T]
}

type pool[T any] struct {
	lock     sync.Mutex
	subPools [64]map[*T]struct{}
	stats    *poolStats[T]
}

// StartStats enables the recording of allocations, reuse and other statistics.
func (p *pool[T]) StartStats() {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.stats.recording = true
}

// PauseStats disables the recording of allocations, reuse and other statistics.
func (p *pool[T]) PauseStats() {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.stats.recording = false
}

// NewPool returns a pool with a size class for every power of two. maxN is ignored; it is kept for compatibility.
func NewPool[T any](maxN ...int) Pool[T] {
	_ = maxN
	p := new(pool[T])
	p.stats = new(poolStats[T])
	p.stats.allocatedFor = make(map[*T][]uintptr)
	return Pool[T]{p}
}

// Make a new slice of size n, reusing a free existing one if possible.
func (p *pool[T]) Make(n int) []T {
	if n <= 0 {
		return nil
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	poolI := ceilLog2(n)
	var ptr *T
	for k := range p.subPools[poolI] {
		ptr = k
		delete(p.subPools[poolI], k)
		break
	}
	if ptr == nil {
		if p.stats.recording {
			p.stats.subPools[poolI].Allocated++
		}
		p.lock.Unlock()
		ptr = getDataPointer(make([]T, 0, 1<<poolI))
		p.lock.Lock()
	}

	if p.stats.recording {
		s := &p.stats.subPools[poolI]

		s.Used++
		s.InUse++
		if n > s.GreatestNUsed {
			s.GreatestNUsed = n
		}
		if s.SmallestNUsed == 0 || s.SmallestNUsed > n {
			s.SmallestNUsed = n
		}

		pcs := make([]uintptr, 2)
		p.stats.allocatedFor[ptr] = pcs[:runtime.Callers(2, pcs)]
	}

	return unsafe.Slice(ptr, 1<<poolI)[:n]
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

		if _, ok := p.subPools[poolI][ptr]; ok {
			panic("slice double-dumped")
		}
		if p.subPools[poolI] == nil {
			p.subPools[poolI] = make(map[*T]struct{})
		}
		p.subPools[poolI][ptr] = struct{}{}

		if p.stats.recording {
			p.stats.subPools[poolI].InUse--
			delete(p.stats.allocatedFor, ptr)
		}
	}
}

func printFrame(frame runtime.Frame) {
	fmt.Printf("\t%s line %d, function %s\n", frame.File, frame.Line, frame.Function)
}

func (p *pool[T]) printInUse() {
	fmt.Println("slices never dumped allocated at:")
	for _, pcs := range p.stats.allocatedFor {
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

// ceilLog2 returns the smallest k such that 1<<k >= n. It requires n >= 1.
func ceilLog2(n int) int {
	return bits.Len(uint(n - 1))
}

func getDataPointer[T any](slice []T) *T {
	return unsafe.SliceData(slice)
}

func (p *pool[T]) PrintPoolStats() {
	p.lock.Lock()
	defer p.lock.Unlock()

	InUse := 0
	subStats := make(map[int]sizedPoolStats)
	for i, s := range p.stats.subPools {
		if s == (sizedPoolStats{}) {
			continue
		}
		s.ReuseRate = float64(s.Used) / float64(max(1, s.Allocated))
		subStats[1<<i] = s
		InUse += s.InUse
	}

	stats := struct {
		SubPools map[int]sizedPoolStats
		InUse    int
	}{subStats, InUse}
	serialized, _ := json.MarshalIndent(stats, "", "  ")
	fmt.Println(string(serialized))
	p.printInUse()
}

func (p *pool[T]) Clone(slice []T) []T {
	res := p.Make(len(slice))
	copy(res, slice)
	return res
}
