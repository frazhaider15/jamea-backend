package models

import (
	"fmt"
	"strings"
)

// MasoolLocation records the geography a masool was responsible for during a
// span of months. Because a masool can be reassigned to a different area over
// time, geography cannot live solely on the (mutable) Masool record — otherwise
// reassigning a masool would silently rewrite the location of all of its past
// reports. Each row is an effective-dated period:
//
//	EffectiveFrom <= month <= EffectiveTo   (inclusive)
//
// Months are "YYYY_MM" strings, which compare correctly lexicographically.
// EffectiveFrom == "" means "from the beginning of time" (used for backfilled
// rows whose true start is unknown); EffectiveTo == "" means the period is
// still open (current).
type MasoolLocation struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	MasoolID      uint   `gorm:"not null;index" json:"masool_id"`
	Module        Module `gorm:"index" json:"module"`
	Province      string `json:"province"`
	Division      string `json:"division"`
	District      string `json:"district"`
	Tehsil        string `json:"tehsil"`
	Area          string `json:"area"`
	EffectiveFrom string `gorm:"index;default:''" json:"effective_from"`
	EffectiveTo   string `gorm:"index;default:''" json:"effective_to"`
}

// LocationFrom extracts the five geographic levels from a masool's flat Data
// into a MasoolLocation (period fields left unset). Key matching is
// case-insensitive; missing levels become empty strings.
func LocationFrom(m Masool) MasoolLocation {
	get := func(key string) string {
		for _, d := range m.Data {
			if strings.EqualFold(strings.TrimSpace(d.Key), key) {
				if s, ok := d.Val.(string); ok {
					return strings.TrimSpace(s)
				}
				return strings.TrimSpace(fmt.Sprintf("%v", d.Val))
			}
		}
		return ""
	}
	return MasoolLocation{
		MasoolID: m.ID,
		Module:   m.Module,
		Province: get("Province"),
		Division: get("Division"),
		District: get("District"),
		Tehsil:   get("Tehsil"),
		Area:     get("Area"),
	}
}
