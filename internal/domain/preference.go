package domain

// Language codes
const (
	LanguageEnglish   = "en"
	LanguageHungarian = "hu"
)

// Preference stores user-specific settings.
type Preference struct {
	PersonID      int16
	Language      string // "en", "hu"
	Theme         bool   // false=dark, true=light
	ActiveColumns any    // jsonb -> map[string]any or typed struct
}

// PreferenceUpdate contains optional fields for partial updates.
type PreferenceUpdate struct {
	Language *string          // "en", "hu"
	Theme    *bool            // false=dark, true=light
	Columns  map[string]bool  // column visibility settings
}
