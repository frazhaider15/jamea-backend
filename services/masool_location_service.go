package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jamea/db"
	"github.com/jamea/models"
	"gorm.io/gorm"
)

// prevMonth returns the "YYYY_MM" month immediately before m. It returns m
// unchanged if m is not in the expected format.
func prevMonth(m string) string {
	parts := strings.Split(m, "_")
	if len(parts) != 2 {
		return m
	}
	year, err1 := strconv.Atoi(parts[0])
	month, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return m
	}
	month--
	if month == 0 {
		month = 12
		year--
	}
	return fmt.Sprintf("%04d_%02d", year, month)
}

// sameGeo reports whether two locations describe the same geography (period
// fields are ignored). Comparison is case-insensitive and trims whitespace.
func sameGeo(a, b models.MasoolLocation) bool {
	eq := func(x, y string) bool {
		return strings.EqualFold(strings.TrimSpace(x), strings.TrimSpace(y))
	}
	return eq(a.Province, b.Province) &&
		eq(a.Division, b.Division) &&
		eq(a.District, b.District) &&
		eq(a.Tehsil, b.Tehsil) &&
		eq(a.Area, b.Area)
}

// locationGeoMap turns a location into a lowercase level -> value map matching
// geoHierarchy ("province", "division", "district", "tehsil", "area").
func locationGeoMap(loc models.MasoolLocation) map[string]string {
	return map[string]string{
		"province": strings.TrimSpace(loc.Province),
		"division": strings.TrimSpace(loc.Division),
		"district": strings.TrimSpace(loc.District),
		"tehsil":   strings.TrimSpace(loc.Tehsil),
		"area":     strings.TrimSpace(loc.Area),
	}
}

// RecordMasoolLocation captures masool's current geography into its location
// history, effective from the current month. Behaviour:
//   - first ever record: opens a period from the current month;
//   - geography unchanged from the open period: no-op;
//   - changed within the same month the open period started: rewrites it in place;
//   - changed in a later month: closes the open period at the previous month and
//     opens a new one from the current month, so past reports keep their old
//     geography while new reports get the new one.
func RecordMasoolLocation(masool models.Masool) error {
	newLoc := models.LocationFrom(masool)
	currentMonth := time.Now().Format("2006_01")

	var open models.MasoolLocation
	err := db.DB.Where("masool_id = ? AND effective_to = ?", masool.ID, "").
		Order("effective_from desc").
		First(&open).Error

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		newLoc.EffectiveFrom = currentMonth
		newLoc.EffectiveTo = ""
		return db.DB.Create(&newLoc).Error
	case err != nil:
		return err
	}

	if sameGeo(open, newLoc) {
		return nil
	}

	// A correction made within the same month the period started just rewrites
	// it; it does not represent a real reassignment, so no new period is opened.
	if open.EffectiveFrom == currentMonth {
		open.Province = newLoc.Province
		open.Division = newLoc.Division
		open.District = newLoc.District
		open.Tehsil = newLoc.Tehsil
		open.Area = newLoc.Area
		return db.DB.Save(&open).Error
	}

	// Real reassignment: close the current period at the previous month so every
	// already-filed report keeps the geography it was filed under, then open a
	// new period from the current month for reports filed going forward.
	open.EffectiveTo = prevMonth(currentMonth)
	if err := db.DB.Save(&open).Error; err != nil {
		return err
	}
	newLoc.EffectiveFrom = currentMonth
	newLoc.EffectiveTo = ""
	return db.DB.Create(&newLoc).Error
}

// resolveMasoolGeo returns the geography (lowercase level -> value) effective
// for the given month, picking the location period that covers it. When no
// period covers the month — e.g. a masool with no recorded history yet — it
// falls back to the masool's current Data so callers still get a best-effort
// answer.
func resolveMasoolGeo(locs []models.MasoolLocation, month string, fallback models.Masool) map[string]string {
	var best models.MasoolLocation
	found := false
	for _, l := range locs {
		if l.EffectiveFrom != "" && l.EffectiveFrom > month {
			continue // not yet effective for this month
		}
		if l.EffectiveTo != "" && month > l.EffectiveTo {
			continue // period already ended before this month
		}
		// Prefer the period with the latest start among those that cover month.
		if !found || l.EffectiveFrom > best.EffectiveFrom {
			best = l
			found = true
		}
	}
	if found {
		return locationGeoMap(best)
	}
	return locationGeoMap(models.LocationFrom(fallback))
}
