package models

import "testing"

func TestFamilyForModelID(t *testing.T) {
	fam, variant, ok := FamilyForModelID("ggml-small.en")
	if !ok || fam.Key != "small" || variant.Label != "English only" {
		t.Fatalf("got %+v %+v ok=%v", fam, variant, ok)
	}
	_, _, ok = FamilyForModelID("not-a-model")
	if ok {
		t.Fatal("expected miss")
	}
}
