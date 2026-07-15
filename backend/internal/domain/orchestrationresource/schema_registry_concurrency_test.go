package orchestrationresource

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func concurrencySchema(factory func() any) Schema {
	return Schema{
		NewSpec: factory,
		Validate: func(Metadata, any) error {
			return nil
		},
	}
}

func concurrentMeta(kind string) TypeMeta {
	return TypeMeta{APIVersion: APIVersionV1Alpha1, Kind: kind}
}

func TestRegistryConcurrentRegisterSameType(t *testing.T) {
	const workers = 32
	registry := NewRegistry()
	start := make(chan struct{})
	results := make(chan error, workers)
	var wait sync.WaitGroup

	for range workers {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			results <- registry.Register(registryMeta, concurrencySchema(func() any {
				return &registrySpec{}
			}))
		}()
	}
	close(start)
	wait.Wait()
	close(results)

	successes := 0
	for err := range results {
		if err == nil {
			successes++
			continue
		}
		require.ErrorIs(t, err, ErrDuplicateSchema)
	}
	require.Equal(t, 1, successes)
}

func TestRegistryConcurrentRegisterAndDecode(t *testing.T) {
	registry := NewRegistry()
	require.NoError(t, registry.Register(registryMeta, validatingRegistrySchema()))
	start := make(chan struct{})
	results := make(chan error, 32)
	var wait sync.WaitGroup

	for index := range 16 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			meta := concurrentMeta(fmt.Sprintf("ConcurrentType%d", index))
			results <- registry.Register(meta, concurrencySchema(func() any {
				return &registrySpec{}
			}))
		}()
	}
	for range 16 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			for range 20 {
				_, err := registry.DecodeAndValidate(validRegistryManifest())
				if err != nil {
					results <- err
					return
				}
			}
			results <- nil
		}()
	}
	close(start)
	wait.Wait()
	close(results)

	for err := range results {
		require.NoError(t, err)
	}
}

func TestRegistryConcurrentDecodeCreatesIndependentResults(t *testing.T) {
	const workers = 32
	shared := &registrySpec{}
	registry := NewRegistry()
	require.NoError(t, registry.Register(registryMeta, concurrencySchema(func() any {
		return shared
	})))
	start := make(chan struct{})
	results := make(chan any, workers)
	errs := make(chan error, workers)
	var wait sync.WaitGroup

	for range workers {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			decoded, err := registry.DecodeAndValidate(validRegistryManifest())
			results <- decoded
			errs <- err
		}()
	}
	close(start)
	wait.Wait()
	close(results)
	close(errs)

	for err := range errs {
		require.NoError(t, err)
	}
	seen := make(map[*registrySpec]struct{}, workers)
	for result := range results {
		spec := result.(*registrySpec)
		_, duplicate := seen[spec]
		require.False(t, duplicate)
		seen[spec] = struct{}{}
	}
	require.Len(t, seen, workers)
	require.NotContains(t, seen, shared)
}
