package services

import (
	"testing"

	"github.com/jamea/models"
)

func TestPrevMonth(t *testing.T) {
	cases := map[string]string{
		"2026_06": "2026_05",
		"2026_01": "2025_12",
		"2026_12": "2026_11",
		"garbage": "garbage",
	}
	for in, want := range cases {
		if got := prevMonth(in); got != want {
			t.Errorf("prevMonth(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestResolveMasoolGeo_Reassignment reproduces the reported bug: Zafar served
// "River Garden" through May 2026 and was reassigned to "Alipur" in June. His
// past reports must still resolve to River Garden, while June onward resolves to
// Alipur.
func TestResolveMasoolGeo_Reassignment(t *testing.T) {
	locs := []models.MasoolLocation{
		{MasoolID: 1, Area: "River Garden", EffectiveFrom: "", EffectiveTo: "2026_05"},
		{MasoolID: 1, Area: "Alipur", EffectiveFrom: "2026_06", EffectiveTo: ""},
	}
	// fallback carries the (now reassigned) current geography.
	fallback := models.Masool{ID: 1, Data: []models.MasoolData{{Key: "Area", Val: "Alipur"}}}

	cases := map[string]string{
		"2026_03": "River Garden",
		"2026_05": "River Garden",
		"2026_06": "Alipur",
		"2026_07": "Alipur",
	}
	for month, want := range cases {
		if got := resolveMasoolGeo(locs, month, fallback)["area"]; got != want {
			t.Errorf("resolveMasoolGeo month=%s area = %q, want %q", month, got, want)
		}
	}
}

// When no period covers the month, resolution falls back to the masool's
// current Data.
func TestResolveMasoolGeo_Fallback(t *testing.T) {
	fallback := models.Masool{ID: 2, Data: []models.MasoolData{{Key: "Area", Val: "Somewhere"}}}
	if got := resolveMasoolGeo(nil, "2026_03", fallback)["area"]; got != "Somewhere" {
		t.Errorf("fallback area = %q, want %q", got, "Somewhere")
	}
}

func TestGeoMapMatchesHierarchy(t *testing.T) {
	provinceOnly := map[string]string{"province": "Punjab", "division": "", "district": "", "tehsil": "", "area": ""}
	withDistrict := map[string]string{"province": "Punjab", "division": "Lahore_Div", "district": "Lahore", "tehsil": "", "area": ""}

	tests := []struct {
		name    string
		geo     map[string]string
		filters map[string]string
		want    bool
	}{
		{"province match, lower empty", provinceOnly, map[string]string{"Province": "Punjab"}, true},
		{"province match but lower populated", withDistrict, map[string]string{"Province": "Punjab"}, false},
		{"district match, lower empty", withDistrict, map[string]string{"District": "Lahore"}, true},
		{"district match case-insensitive", withDistrict, map[string]string{"District": "lahore"}, true},
		{"wrong value", provinceOnly, map[string]string{"Province": "Sindh"}, false},
		{"no filter", provinceOnly, map[string]string{}, true},
	}
	for _, tc := range tests {
		if got := geoMapMatchesHierarchy(tc.geo, tc.filters); got != tc.want {
			t.Errorf("%s: geoMapMatchesHierarchy = %v, want %v", tc.name, got, tc.want)
		}
	}
}
