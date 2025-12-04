package umbra_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/robert.malhotra/go-sar-vendor/pkg/umbra"
)

func TestGetRestrictedAccessAreas(t *testing.T) {
	expected := []umbra.RestrictedAccessArea{
		{
			ID:   "raa-1",
			Name: "Restricted Zone 1",
		},
		{
			ID:   "raa-2",
			Name: "Restricted Zone 2",
		},
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/tasking/restricted-access-areas")
		requireAuth(t, r, "test-token")
		jsonResponse(w, http.StatusOK, expected)
	})

	areas, err := cli.GetRestrictedAccessAreas(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(areas) != 2 {
		t.Errorf("expected 2 areas, got %d", len(areas))
	}
	if areas[0].Name != "Restricted Zone 1" {
		t.Errorf("expected name 'Restricted Zone 1', got %s", areas[0].Name)
	}
}

func TestGetOrganizationSettings(t *testing.T) {
	expected := umbra.OrganizationSettings{
		ID:                    "settings-123",
		OrganizationID:        "org-456",
		DefaultDeliveryConfig: "dc-789",
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/admin/organization-settings")
		requireAuth(t, r, "test-token")
		jsonResponse(w, http.StatusOK, expected)
	})

	settings, err := cli.GetOrganizationSettings(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if settings.OrganizationID != expected.OrganizationID {
		t.Errorf("expected org ID %s, got %s", expected.OrganizationID, settings.OrganizationID)
	}
}

func TestListProductConstraintsAdmin(t *testing.T) {
	expected := []umbra.ProductConstraint{
		{
			ProductType:       "GEC",
			SceneSize:         "5km",
			MinGrazingDegrees: 25.0,
			MaxGrazingDegrees: 70.0,
		},
	}

	cli, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requireMethod(t, r, http.MethodGet)
		requirePath(t, r, "/admin/product-constraints")
		requireAuth(t, r, "test-token")
		jsonResponse(w, http.StatusOK, expected)
	})

	constraints, err := cli.ListProductConstraintsAdmin(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(constraints) != 1 {
		t.Errorf("expected 1 constraint, got %d", len(constraints))
	}
}
