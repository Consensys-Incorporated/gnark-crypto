// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package utils

import (
	"encoding/json"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"unsafe"
)

// Memory management for polynomials
// The pool statistics are guarded by a mutex; Make and Dump may be called concurrently.
// TODO: There is a lot of "unsafe" memory management here and needs to be vetted thoroughly

type sizedPool[T any] struct {
	maxN  int
	pool  sync.Pool
	stats poolStats
}

type inUseData[T any] struct {
	allocatedFor []uintptr
	pool         *sizedPool[T]
}

// Pool is a pool of slices of T, grouped by maximum length.
type Pool[T any] struct {
	//lock     sync.Mutex
	inUse    sync.Map
	subPools []sizedPool[T]
}

func (p *sizedPool[T]) get(n int) *T {
	p.stats.make(n)
	return p.pool.Get().(*T)
}

func (p *sizedPool[T]) put(ptr *T) {
	p.stats.dump()
	p.pool.Put(ptr)
}

// NewPool returns a Pool whose sub-pools hold slices of capacity up to each of maxN.
func NewPool[T any](maxN ...int) (pool Pool[T]) {

	sort.Ints(maxN)
	pool = Pool[T]{
		subPools: make([]sizedPool[T], len(maxN)),
	}

	for i := range pool.subPools {
		subPool := &pool.subPools[i]
		subPool.maxN = maxN[i]
		subPool.pool = sync.Pool{
			New: func() any {
				subPool.stats.allocated()
				return getDataPointer(make([]T, 0, subPool.maxN))
			},
		}
	}
	return
}

func (p *Pool[T]) findCorrespondingPool(n int) *sizedPool[T] {
	poolI := 0
	for poolI < len(p.subPools) && n > p.subPools[poolI].maxN {
		poolI++
	}
	return &p.subPools[poolI] // out of bounds error here would mean that n is too large
}

func (p *Pool[T]) Make(n int) []T {
	pool := p.findCorrespondingPool(n)
	ptr := pool.get(n)
	p.addInUse(ptr, pool)
	return unsafe.Slice(ptr, n)
}

// Dump dumps a set of polynomials into the pool
func (p *Pool[T]) Dump(slices ...[]T) {
	for _, slice := range slices {
		ptr := getDataPointer(slice)
		if metadata, ok := p.inUse.Load(ptr); ok {
			p.inUse.Delete(ptr)
			metadata.(inUseData[T]).pool.put(ptr)
		} else {
			panic("attempting to dump a slice not created by the pool")
		}
	}
}

func (p *Pool[T]) addInUse(ptr *T, pool *sizedPool[T]) {
	pcs := make([]uintptr, 2)
	n := runtime.Callers(3, pcs)

	if prevPcs, ok := p.inUse.Load(ptr); ok { // TODO: remove if unnecessary for security
		panic(fmt.Errorf("re-allocated non-dumped slice, previously allocated at %v", runtime.CallersFrames(prevPcs.(inUseData[T]).allocatedFor)))
	}
	p.inUse.Store(ptr, inUseData[T]{
		allocatedFor: pcs[:n],
		pool:         pool,
	})
}

func printFrame(frame runtime.Frame) {
	fmt.Printf("\t%s line %d, function %s\n", frame.File, frame.Line, frame.Function)
}

func (p *Pool[T]) printInUse() {
	fmt.Println("slices never dumped allocated at:")
	p.inUse.Range(func(_, pcs any) bool {
		fmt.Println("-------------------------")

		var frame runtime.Frame
		frames := runtime.CallersFrames(pcs.(inUseData[T]).allocatedFor)
		more := true
		for more {
			frame, more = frames.Next()
			printFrame(frame)
		}
		return true
	})
}

type poolStats struct {
	mu            sync.Mutex
	Used          int
	Allocated     int
	ReuseRate     float64
	InUse         int
	GreatestNUsed int
	SmallestNUsed int
}

type poolsStats struct {
	SubPools []poolStats
	InUse    int
}

func (s *poolStats) allocated() {
	s.mu.Lock()
	s.Allocated++
	s.mu.Unlock()
}

// snapshot returns a copy of the statistics with ReuseRate computed.
func (s *poolStats) snapshot() poolStats {
	s.mu.Lock()
	defer s.mu.Unlock()
	return poolStats{
		Used:          s.Used,
		Allocated:     s.Allocated,
		ReuseRate:     float64(s.Used) / float64(s.Allocated),
		InUse:         s.InUse,
		GreatestNUsed: s.GreatestNUsed,
		SmallestNUsed: s.SmallestNUsed,
	}
}

func (s *poolStats) make(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Used++
	s.InUse++
	if n > s.GreatestNUsed {
		s.GreatestNUsed = n
	}
	if s.SmallestNUsed == 0 || s.SmallestNUsed > n {
		s.SmallestNUsed = n
	}
}

func (s *poolStats) dump() {
	s.mu.Lock()
	s.InUse--
	s.mu.Unlock()
}

func getDataPointer[T any](slice []T) *T {
	return unsafe.SliceData(slice)
}

func (p *Pool[T]) PrintPoolStats() {
	InUse := 0
	subStats := make([]poolStats, len(p.subPools))
	for i := range p.subPools {
		subPool := &p.subPools[i]
		subStats[i] = subPool.stats.snapshot()
		InUse += subStats[i].InUse
	}

	stats := poolsStats{
		SubPools: subStats,
		InUse:    InUse,
	}
	serialized, _ := json.MarshalIndent(stats, "", "  ")
	fmt.Println(string(serialized))
	p.printInUse()
}

func (p *Pool[T]) Clone(slice []T) []T {
	res := p.Make(len(slice))
	copy(res, slice)
	return res
}
