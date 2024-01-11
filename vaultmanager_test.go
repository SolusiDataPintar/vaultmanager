package vaultmanager_test

import (
	"sync"
	"testing"
	"time"

	"github.com/SolusiDataPintar/vaultmanager"
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
