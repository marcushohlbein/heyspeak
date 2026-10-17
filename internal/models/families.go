package models

// ModelVariant is one downloadable artifact within a family (e.g. multilingual vs English-only).
type ModelVariant struct {
	Label string
	ID    string
}

// ModelFamily groups variants that share a size tier in the settings UI.
type ModelFamily struct {
	Key          string
	Name         string
	Requirements string
	Variants     []ModelVariant
}

// DefaultFamilies returns catalog tiers in display order.
func DefaultFamilies() []ModelFamily {
	return []ModelFamily{
		{
			Key:          "tiny",
			Name:         "Whisper Tiny",
			Requirements: "~1 GB RAM · fastest, lowest accuracy",
			Variants: []ModelVariant{
				{Label: "Multilingual", ID: "ggml-tiny"},
				{Label: "English only", ID: "ggml-tiny.en"},
			},
		},
		{
			Key:          "base",
			Name:         "Whisper Base",
			Requirements: "~1.5 GB RAM · good balance for testing",
			Variants: []ModelVariant{
				{Label: "Multilingual", ID: "ggml-base"},
				{Label: "English only", ID: "ggml-base.en"},
			},
		},
		{
			Key:          "small",
			Name:         "Whisper Small",
			Requirements: "~2 GB RAM · recommended for daily use",
			Variants: []ModelVariant{
				{Label: "Multilingual", ID: "ggml-small"},
				{Label: "English only", ID: "ggml-small.en"},
			},
		},
		{
			Key:          "medium",
			Name:         "Whisper Medium",
			Requirements: "~5 GB RAM · high accuracy, slower",
			Variants: []ModelVariant{
				{Label: "Multilingual", ID: "ggml-medium"},
				{Label: "English only", ID: "ggml-medium.en"},
			},
		},
		{
			Key:          "large-v3-turbo",
			Name:         "Whisper Large v3 Turbo",
			Requirements: "~6 GB RAM · best quality without full Large v3 size",
			Variants: []ModelVariant{
				{Label: "Multilingual", ID: "ggml-large-v3-turbo"},
			},
		},
		{
			Key:          "large-v3",
			Name:         "Whisper Large v3",
			Requirements: "~10 GB RAM · highest accuracy, large download",
			Variants: []ModelVariant{
				{Label: "Multilingual", ID: "ggml-large-v3"},
			},
		},
	}
}

// FamilyForModelID returns the family and variant label for a catalog model ID.
func FamilyForModelID(id string) (ModelFamily, ModelVariant, bool) {
	for _, fam := range DefaultFamilies() {
		for _, v := range fam.Variants {
			if v.ID == id {
				return fam, v, true
			}
		}
	}
	return ModelFamily{}, ModelVariant{}, false
}
