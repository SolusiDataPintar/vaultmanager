package vaultmanager_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/SolusiDataPintar/vaultmanager"
	"github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

func TestOpen(t *testing.T) {
	err := vaultmanager.Open(nil)
	require.NoError(t, err)
}

func TestGetClient(t *testing.T) {
	err := vaultmanager.Open(nil)
	require.NoError(t, err)

	c := vaultmanager.GetClient()
	require.NotNil(t, c)
}

func TestWriteKVv2(t *testing.T) {
	err := vaultmanager.Open(nil)
	require.NoError(t, err)

	data := map[string]any{
		"test1": "test5",
		"test2": "test4",
		"test3": "test3",
		"test4": "test2",
		"test5": "test1",
	}
	err = vaultmanager.WriteKVv2(context.Background(), "chainsmart", "test/vault-manager-write", data)
	require.NoError(t, err)

	err = vaultmanager.GetClient().KVv2("chainsmart").Delete(context.Background(), "test/vault-manager-write")
	require.NoError(t, err)
}

func TestReadKVv2(t *testing.T) {
	err := vaultmanager.Open(nil)
	require.NoError(t, err)

	data := map[string]any{
		"test1": "test5",
		"test2": "test4",
		"test3": "test3",
		"test4": "test2",
		"test5": "test1",
	}
	err = vaultmanager.WriteKVv2(context.Background(), "chainsmart", "test/vault-manager-read", data)
	require.NoError(t, err)

	secretData, err := vaultmanager.ReadKVv2(context.Background(), "chainsmart", "test/vault-manager-read")
	require.NoError(t, err)
	require.Equal(t, data, secretData)

	err = vaultmanager.GetClient().KVv2("chainsmart").Delete(context.Background(), "test/vault-manager-read")
	require.NoError(t, err)
}

func TestReadKVv2NotFound(t *testing.T) {
	err := vaultmanager.Open(nil)
	require.NoError(t, err)

	secretData, err := vaultmanager.ReadKVv2(context.Background(), "chainsmart", "test/vault-manager-not-found")
	require.EqualError(t, err, api.ErrSecretNotFound.Error())
	require.True(t, errors.Is(err, api.ErrSecretNotFound))
	require.NotNil(t, secretData)
	require.Empty(t, secretData)
}

func TestManageTokenLifecycle(t *testing.T) {
	err := vaultmanager.Open(nil)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err := vaultmanager.ManageTokenLifecycle(ctx)
		require.NoError(t, err)
		wg.Done()
	}()
	wg.Wait()
}
