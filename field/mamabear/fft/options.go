// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fft

import (
	"runtime"

	"github.com/consensys/gnark-crypto/field/mamabear"
)

// Option defines option for altering the behavior of FFT methods.
// See the descriptions of functions returning instances of this type for
// particular options.
type Option func(fftConfig) fftConfig

type fftConfig struct {
	coset   bool
	nbTasks int
}

// OnCoset if provided, FFT(a) returns the evaluation of a on a coset.
func OnCoset() Option {
	return func(opt fftConfig) fftConfig {
		opt.coset = true
		return opt
	}
}

// WithNbTasks sets the max number of task (go routine) to spawn. Must be between 1 and 512.
func WithNbTasks(nbTasks int) Option {
	if nbTasks < 1 {
		nbTasks = 1
	} else if nbTasks > 512 {
		nbTasks = 512
	}
	return func(opt fftConfig) fftConfig {
		opt.nbTasks = nbTasks
		return opt
	}
}

// default options
func fftOptions(opts []Option) fftConfig {
	// apply options
	opt := fftConfig{
		coset:   false,
		nbTasks: runtime.NumCPU(),
	}
	for _, option := range opts {
		opt = option(opt)
	}
	return opt
}

// DomainOption defines option for altering the definition of the FFT domain
// See the descriptions of functions returning instances of this type for
// particular options.
type DomainOption func(*domainConfig)

type domainConfig struct {
	shift          *mamabear.Element
	withPrecompute bool
	withCache      bool
}

// WithShift sets the FrMultiplicativeGen of the domain.
// Default is generator of the largest 2-adic subgroup.
func WithShift(shift mamabear.Element) DomainOption {
	return func(opt *domainConfig) {
		opt.shift = new(mamabear.Element).Set(&shift)
	}
}

// WithoutPrecompute disables precomputation of twiddles in the domain.
// When this option is set, FFTs will be slower, but will use less memory.
func WithoutPrecompute() DomainOption {
	return func(opt *domainConfig) {
		opt.withPrecompute = false
	}
}

// WithCache enables domain caching
func WithCache() DomainOption {
	return func(opt *domainConfig) {
		opt.withCache = true
	}
}

// default options
func domainOptions(opts ...DomainOption) domainConfig {
	// apply options
	opt := domainConfig{
		withPrecompute: true,
		withCache:      false,
	}
	for _, option := range opts {
		option(&opt)
	}
	return opt
}
