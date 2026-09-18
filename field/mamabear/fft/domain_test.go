// Copyright 2020-2026 Consensys Software Inc.
// Licensed under the Apache License, Version 2.0. See the LICENSE file for details.

package fft

import (
	"bytes"
	"reflect"
	"runtime"
	"testing"

	"github.com/consensys/gnark-crypto/field/mamabear"
	"github.com/stretchr/testify/require"
)

func TestDomainSerialization(t *testing.T) {
	domain := NewDomain(1 << 6)
	var reconstructed Domain

	var buf bytes.Buffer
	written, err := domain.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	var read int64
	read, err = reconstructed.ReadFrom(&buf)
	if err != nil {
		t.Fatal(err)
	}

	if written != read {
		t.Fatal("didn't read as many bytes as we wrote")
	}
	if !reflect.DeepEqual(domain, &reconstructed) {
		t.Fatal("Domain.SetBytes(Bytes()) failed")
	}
}

func TestNewDomainCache(t *testing.T) {
	t.Run("CacheWithoutShift", func(t *testing.T) {
		key1 := domainCacheKey{
			m:   256,
			gen: GeneratorFullMultiplicativeGroup(),
		}
		require.Nil(t, getCachedDomain(key1), "cache should be empty initially")
		domain1 := NewDomain(256, WithCache())
		expected1 := NewDomain(256)
		require.Equal(t, domain1, expected1, "domain1 should equal expected1")
		require.Same(t, domain1, getCachedDomain(key1), "domain1 should be stored in cache")
	})

	t.Run("CacheWithShift", func(t *testing.T) {
		shift := mamabear.NewElement(5)
		key2 := domainCacheKey{
			m:   512,
			gen: shift,
		}
		require.Nil(t, getCachedDomain(key2), "cache should be empty initially")
		domain2 := NewDomain(512, WithShift(shift), WithCache())
		expected2 := NewDomain(512, WithShift(shift))
		require.Equal(t, domain2, expected2, "domain2 should equal expected2")
		require.Same(t, domain2, getCachedDomain(key2), "domain2 should be stored in cache")
	})
}

func TestGCBehavior(t *testing.T) {
	t.Run("DomainKeptAliveNotCollected", func(t *testing.T) {
		domain := NewDomain(1<<20, WithCache())
		require.NotNil(t, domain, "domain should not be empty")
		key := domainCacheKey{
			m:   1 << 20,
			gen: GeneratorFullMultiplicativeGroup(),
		}
		require.NotNil(t, getCachedDomain(key), "domain should be cached")
		runtime.GC()
		require.NotNil(t, getCachedDomain(key), "domain should still be cached")
		runtime.KeepAlive(domain)
	})

	t.Run("UnreferencedDomainCollected", func(t *testing.T) {
		domain := NewDomain(1<<19, WithCache())
		key := domainCacheKey{
			m:   1 << 19,
			gen: GeneratorFullMultiplicativeGroup(),
		}
		require.NotNil(t, getCachedDomain(key), "domain should be cached")
		// Last use of domain
		_ = domain.Cardinality
		runtime.GC()
		require.Nil(t, getCachedDomain(key), "unreferenced domain should be collected and removed from cache")
	})
}

func BenchmarkNewDomainCache(b *testing.B) {
	b.Run("WithCache", func(b *testing.B) {
		cached := NewDomain(1<<20, WithCache())
		for b.Loop() {
			_ = NewDomain(1<<20, WithCache())
		}
		runtime.KeepAlive(cached)
	})

	b.Run("WithoutCache", func(b *testing.B) {
		for b.Loop() {
			_ = NewDomain(1 << 20)
		}
	})
}

// Helper functions
func getCachedDomain(key domainCacheKey) *Domain {
	keyMapLock.Lock()
	keyLock, exists := domainGenLocks[key]
	if !exists {
		keyMapLock.Unlock()
		return nil
	}

	keyLock.Lock()
	defer keyLock.Unlock()
	keyMapLock.Unlock()

	domainMapLock.Lock()
	defer domainMapLock.Unlock()
	if weak, exists := domainCache[key]; exists {
		return weak.Value()
	}
	return nil
}
