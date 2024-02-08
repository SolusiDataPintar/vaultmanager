package vaultmanager_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/SolusiDataPintar/vaultmanager"
	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/require"
)

func TestNewVaultManager(t *testing.T) {
	vm, err := vaultmanager.NewVaultManager(nil)
	require.NoError(t, err)
	require.NotNil(t, vm)
}

func TestGetClient(t *testing.T) {
	vm, err := vaultmanager.NewVaultManager(nil)
	require.NoError(t, err)
	require.NotNil(t, vm)

	c := vm.GetClient()
	require.NotNil(t, c)
}

func TestWriteKVv2(t *testing.T) {
	vm, err := vaultmanager.NewVaultManager(nil)
	require.NoError(t, err)
	require.NotNil(t, vm)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	data := map[string]any{
		"test1": "test5",
		"test2": "test4",
		"test3": "test3",
		"test4": "test2",
		"test5": "test1",
	}
	err = vm.WriteKVv2(ctx, "chainsmart", "test/vault-manager-write", data)
	require.NoError(t, err)

	err = vm.GetClient().KVv2("chainsmart").Delete(ctx, "test/vault-manager-write")
	require.NoError(t, err)
}

func TestReadKVv2(t *testing.T) {
	vm, err := vaultmanager.NewVaultManager(nil)
	require.NoError(t, err)
	require.NotNil(t, vm)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	data := map[string]any{
		"test1": "test5",
		"test2": "test4",
		"test3": "test3",
		"test4": "test2",
		"test5": "test1",
	}
	err = vm.WriteKVv2(ctx, "chainsmart", "test/vault-manager-read", data)
	require.NoError(t, err)

	secretData, err := vm.ReadKVv2(ctx, "chainsmart", "test/vault-manager-read")
	require.NoError(t, err)
	require.Equal(t, data, secretData)

	err = vm.GetClient().KVv2("chainsmart").Delete(ctx, "test/vault-manager-read")
	require.NoError(t, err)
}

func TestReadKVv2NotFound(t *testing.T) {
	vm, err := vaultmanager.NewVaultManager(nil)
	require.NoError(t, err)
	require.NotNil(t, vm)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	secretData, err := vm.ReadKVv2(ctx, "chainsmart", "test/vault-manager-not-found")
	require.EqualError(t, err, api.ErrSecretNotFound.Error())
	require.True(t, errors.Is(err, api.ErrSecretNotFound))
	require.NotNil(t, secretData)
	require.Empty(t, secretData)
}

func TestManageTokenLifecycle(t *testing.T) {
	vm, err := vaultmanager.NewVaultManager(nil)
	require.NoError(t, err)
	require.NotNil(t, vm)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err := vm.ManageTokenLifecycle(ctx)
		require.NoError(t, err)
		wg.Done()
	}()
	wg.Wait()
}
