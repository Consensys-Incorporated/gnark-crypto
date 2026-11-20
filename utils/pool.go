// Copyright 2020-2026 Consensys Incorporated
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package utils

import (
	"cmp"
	"errors"
	"fmt"
	"math/bits"
	"runtime"
	"slices"
	"strings"
	"sync"
)

// callSiteDepth is the number of frames recorded for each slice handed out by Make while tracking undumped slices.
const callSiteDepth = 8

// callSite is the call stack of a call to Make, as program counters. Unused trailing entries are 0.
type callSite [callSiteDepth]uintptr

// Pool is a pool of slices of T, grouped by power-of-two capacity. Its copies share the same slices.
// Dumped slices are retained for the lifetime of the Pool and are not evicted by the garbage collector,
// so a Pool should be scoped to an operation rather than kept globally.
type Pool[T any] struct {
	*pool[T]
}

type pool[T any] struct {
	lock             sync.Mutex
	subPools         [bits.UintSize][][]T
	trackingUndumped bool
	undumped         map[*T]callSite // call stack of Make for each tracked slice not yet dumped, keyed by data pointer
}

// StartTrackingUndumped starts recording the call site of every slice handed out by Make, until the slice is
// dumped, so that CheckUndumped can report slices that were never dumped. It is a development aid for finding
// missing calls to Dump.
func (p *pool[T]) StartTrackingUndumped() {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.trackingUndumped = true
}

// PauseTrackingUndumped stops recording slices handed out by Make. Slices already recorded are still removed
// when dumped.
func (p *pool[T]) PauseTrackingUndumped() {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.trackingUndumped = false
}

// NewPool returns a pool with a size class for every power of two. maxN is ignored; it is kept for compatibility.
func NewPool[T any](maxN ...int) Pool[T] {
	_ = maxN
	p := new(pool[T])
	p.undumped = make(map[*T]callSite)
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
		if p.trackingUndumped {
			p.trackMake(slice)
		}
		p.lock.Unlock()
		return slice[:n]
	}

	trackingUndumped := p.trackingUndumped
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
	var site callSite
	runtime.Callers(3, site[:])
	p.undumped[getDataPointer(slice)] = site
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

// ceilLog2 returns the smallest k such that 1<<k >= n. It requires n >= 1.
func ceilLog2(n int) int {
	return bits.Len(uint(n - 1))
}

// getDataPointer returns a pointer to the first element of slice's underlying array. It requires cap(slice) ≥ 1.
func getDataPointer[T any](slice []T) *T {
	return &slice[:1][0]
}

// CheckUndumped returns an error listing, by call site, the slices handed out by Make while tracking that have
// not been dumped, or nil if there are none. Call it where every slice from Make should have been dumped.
func (p *pool[T]) CheckUndumped() error {
	p.lock.Lock()
	defer p.lock.Unlock()

	if len(p.undumped) == 0 {
		return nil
	}

	counts := make(map[callSite]int)
	for _, site := range p.undumped {
		counts[site]++
	}

	type siteReport struct {
		count  int
		frames string
	}
	reports := make([]siteReport, 0, len(counts))
	for site, count := range counts {
		n := slices.Index(site[:], 0)
		if n < 0 {
			n = len(site)
		}
		var frames strings.Builder
		callersFrames := runtime.CallersFrames(site[:n])
		for more := true; more; {
			var frame runtime.Frame
			frame, more = callersFrames.Next()
			fmt.Fprintf(&frames, "\t%s line %d, function %s\n", frame.File, frame.Line, frame.Function)
		}
		reports = append(reports, siteReport{count, frames.String()})
	}
	slices.SortFunc(reports, func(a, b siteReport) int {
		return cmp.Or(cmp.Compare(b.count, a.count), strings.Compare(a.frames, b.frames))
	})

	var msg strings.Builder
	fmt.Fprintf(&msg, "%d slices from Make were never dumped, from %d call sites:\n", len(p.undumped), len(reports))
	for _, report := range reports {
		fmt.Fprintf(&msg, "%d made at:\n%s", report.count, report.frames)
	}
	return errors.New(strings.TrimSuffix(msg.String(), "\n"))
}

func (p *pool[T]) Clone(slice []T) []T {
	res := p.Make(len(slice))
	copy(res, slice)
	return res
}
