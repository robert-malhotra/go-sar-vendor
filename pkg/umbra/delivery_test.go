package umbra_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/robert.malhotra/go-sar-vendor/pkg/umbra"
)

func TestCreateDeliveryConfig(t *testing.T) {
	expected := umbra.DeliveryConfig{
		ID:     "dc-123",
		Name:   "My S3 Bucket",
		Type:   umbra.DeliveryTypeS3UmbraRole,
		Status: umbra.DeliveryConfigStatusUnverified,
		Bucket: "my-bucket",
		Path:   "/data",
		Region: "us-west-2",
		CreatedAt: time.Now(),
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodPost)
		requirePath(t, r, "/delivery/delivery-config")
		requireAuth(t, r, "test-token")
		requireContentType(t, r, "application/json")
		jsonResponse(w, http.StatusCreated, expected)
	})

	req := &umbra.CreateDeliveryConfigRequest{
		Name:   "My S3 Bucket",
		Type:   umbra.DeliveryTypeS3UmbraRole,
		Bucket: "my-bucket",
		Path:   "/data",
		Region: "us-west-2",
	}

	dc, err := cli.CreateDeliveryConfig(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dc.ID != expected.ID {
		t.Errorf("expected ID %s, got %s", expected.ID, dc.ID)
	}
	if dc.Status != expected.Status {
		t.Errorf("expected status %s, got %s", expected.Status, dc.Status)
	}
}

func TestGetDeliveryConfig(t *testing.T) {
	expected := umbra.DeliveryConfig{
		ID:         "dc-456",
		Name:       "GCP Bucket",
		Type:       umbra.DeliveryTypeGCPWIF,
		Status:     umbra.DeliveryConfigStatusActive,
		ProjectID:  "my-project",
		BucketName: "my-gcp-bucket",
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/delivery/delivery-config/dc-456")
		requireAuth(t, r, "test-token")
		jsonResponse(w, http.StatusOK, expected)
	})

	dc, err := cli.GetDeliveryConfig(context.Background(), "dc-456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dc.ID != expected.ID {
		t.Errorf("expected ID %s, got %s", expected.ID, dc.ID)
	}
	if dc.Type != expected.Type {
		t.Errorf("expected type %s, got %s", expected.Type, dc.Type)
	}
}

func TestListDeliveryConfigs(t *testing.T) {
	expected := []umbra.DeliveryConfig{
		{ID: "dc-1", Name: "Config 1", Status: umbra.DeliveryConfigStatusActive},
		{ID: "dc-2", Name: "Config 2", Status: umbra.DeliveryConfigStatusUnverified},
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/delivery/delivery-configs")
		requireAuth(t, r, "test-token")
		jsonResponse(w, http.StatusOK, expected)
	})

	configs, err := cli.ListDeliveryConfigs(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(configs) != 2 {
		t.Errorf("expected 2 configs, got %d", len(configs))
	}
}

func TestVerifyDeliveryConfig(t *testing.T) {
	expected := umbra.DeliveryConfig{
		ID:     "dc-101",
		Name:   "Verified Config",
		Status: umbra.DeliveryConfigStatusActive,
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodPost)
		requirePath(t, r, "/delivery/delivery-config/verify")
		requireAuth(t, r, "test-token")
		jsonResponse(w, http.StatusOK, expected)
	})

	dc, err := cli.VerifyDeliveryConfig(context.Background(), "dc-101")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dc.Status != umbra.DeliveryConfigStatusActive {
		t.Errorf("expected status %s, got %s", umbra.DeliveryConfigStatusActive, dc.Status)
	}
}

func TestDeleteDeliveryConfig(t *testing.T) {
	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodDelete)
		requirePath(t, r, "/delivery/delivery-config/dc-202")
		requireAuth(t, r, "test-token")
		w.WriteHeader(http.StatusNoContent)
	})

	err := cli.DeleteDeliveryConfig(context.Background(), "dc-202")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetDeliveryConfig_NotFound(t *testing.T) {
	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		errorResponse(w, http.StatusNotFound, "Delivery config not found")
	})

	_, err := cli.GetDeliveryConfig(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !umbra.IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestCreateDeliveryConfig_Error(t *testing.T) {
	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		errorResponse(w, http.StatusBadRequest, "Invalid configuration")
	})

	req := &umbra.CreateDeliveryConfigRequest{
		Name: "Bad Config",
	}

	_, err := cli.CreateDeliveryConfig(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !umbra.IsBadRequest(err) {
		t.Errorf("expected bad request error, got %v", err)
	}
}

func TestGetCollectMetadataSchema(t *testing.T) {
	expected := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"umbra:collect_id": map[string]interface{}{"type": "string"},
		},
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/delivery/collect-metadata/schema")
		requireAuth(t, r, "test-token")
		jsonResponse(w, http.StatusOK, expected)
	})

	schema, err := cli.GetCollectMetadataSchema(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema["type"] != "object" {
		t.Errorf("expected schema type object, got %v", schema["type"])
	}
}
