package vaultmanager_test

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/SolusiDataPintar/vaultmanager"
	"github.com/hashicorp/vault/api"
)

func TestWriteKVv2(t *testing.T) {
	vm, err := vaultmanager.NewVaultManager(nil)
	if err != nil {
		t.Fatalf("failed to create vault manager: %s", err)
	}
	if vm == nil {
		t.Fatal("vault manager is nil")
	}

	data := map[string]any{
		"test1": "test5",
		"test2": "test4",
		"test3": "test3",
		"test4": "test2",
		"test5": "test1",
	}
	err = vm.WriteKVv2(t.Context(), "dev-test", "test/vault-manager-write", data)
	if err != nil {
		t.Fatalf("failed to write KVv2: %s", err)
	}

	err = vm.GetClient().KVv2("dev-test").Delete(t.Context(), "test/vault-manager-write")
	if err != nil {
		t.Fatalf("failed to delete KVv2: %s", err)
	}
}

func TestReadKVv2(t *testing.T) {
	vm, err := vaultmanager.NewVaultManager(nil)
	if err != nil {
		t.Fatalf("failed to create vault manager: %s", err)
	}
	if vm == nil {
		t.Fatal("vault manager is nil")
	}

	data := map[string]any{
		"test1": "test5",
		"test2": "test4",
		"test3": "test3",
		"test4": "test2",
		"test5": "test1",
	}
	err = vm.WriteKVv2(t.Context(), "dev-test", "test/vault-manager-read", data)
	if err != nil {
		t.Fatalf("failed to write KVv2: %s", err)
	}

	secretData, err := vm.ReadKVv2(t.Context(), "dev-test", "test/vault-manager-read")
	if err != nil {
		t.Fatalf("failed to read KVv2: %s", err)
	}
	if !reflect.DeepEqual(data, secretData) {
		t.Fatalf("expected %v, got %v", data, secretData)
	}

	err = vm.GetClient().KVv2("dev-test").Delete(t.Context(), "test/vault-manager-read")
	if err != nil {
		t.Fatalf("failed to delete KVv2: %s", err)
	}
}

func TestReadKVv2NotFound(t *testing.T) {
	vm, err := vaultmanager.NewVaultManager(nil)
	if err != nil {
		t.Fatalf("failed to create vault manager: %s", err)
	}
	if vm == nil {
		t.Fatal("vault manager is nil")
	}

	secretData, err := vm.ReadKVv2(t.Context(), "dev-test", "test/vault-manager-not-found")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	// Check if the error is the expected one
	if !errors.Is(err, api.ErrSecretNotFound) {
		t.Fatalf("expected error %v, got %v", api.ErrSecretNotFound, err)
	}
	if secretData == nil {
		t.Fatal("expected secretData to be non-nil, got nil")
	}
	if len(secretData) != 0 {
		t.Fatalf("expected secretData to be empty, got %v", secretData)
	}
}

func TestDeleteKVv2(t *testing.T) {
	vm, err := vaultmanager.NewVaultManager(nil)
	if err != nil {
		t.Fatalf("failed to create vault manager: %s", err)
	}
	if vm == nil {
		t.Fatal("vault manager is nil")
	}

	data := map[string]any{
		"test1": "test5",
		"test2": "test4",
		"test3": "test3",
		"test4": "test2",
		"test5": "test1",
	}
	err = vm.WriteKVv2(t.Context(), "dev-test", "test/vault-manager-write", data)
	if err != nil {
		t.Fatalf("failed to write KVv2: %s", err)
	}

	err = vm.DeleteKVv2(t.Context(), "dev-test", "test/vault-manager-write")
	if err != nil {
		t.Fatalf("failed to delete KVv2: %s", err)
	}
}

func TestManageTokenLifecycle(t *testing.T) {
	vm, err := vaultmanager.NewVaultManager(nil)
	if err != nil {
		t.Fatalf("failed to create vault manager: %s", err)
	}
	if vm == nil {
		t.Fatal("vault manager is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		err := vm.ManageTokenLifecycle(ctx)
		if err != nil {
			t.Errorf("failed to manage token lifecycle: %s", err)
		}
		wg.Done()
	}()
	wg.Wait()
}
