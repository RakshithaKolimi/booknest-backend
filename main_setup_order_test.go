package main

import "testing"

func TestLoadOrderServiceRuntimeConfigDefaultsToMonolith(t *testing.T) {
	cfg, err := loadOrderServiceRuntimeConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.UseMicroservice {
		t.Fatal("expected monolith to be the default")
	}
}

func TestLoadOrderServiceRuntimeConfigSupportsMode(t *testing.T) {
	t.Setenv("ORDER_SERVICE_MODE", "microservice")
	t.Setenv("ORDER_SERVICE_GRPC_ADDR", "localhost:50051")

	cfg, err := loadOrderServiceRuntimeConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !cfg.UseMicroservice {
		t.Fatal("expected microservice mode to be enabled")
	}
	if cfg.GRPCAddress != "localhost:50051" {
		t.Fatalf("expected grpc addr to be preserved, got %q", cfg.GRPCAddress)
	}
}

func TestLoadOrderServiceRuntimeConfigRejectsMissingAddr(t *testing.T) {
	t.Setenv("USE_ORDER_MICROSERVICE", "true")

	_, err := loadOrderServiceRuntimeConfig()
	if err == nil {
		t.Fatal("expected missing grpc addr to fail")
	}
}
