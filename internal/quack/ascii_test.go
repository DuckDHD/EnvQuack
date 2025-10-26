package quack

import (
	"strings"
	"testing"
)

func TestGetHappyDuck(t *testing.T) {
	duck := GetHappyDuck()
	if duck == "" {
		t.Errorf("GetHappyDuck() should return non-empty string")
	}
	// Should contain ASCII art pattern
	if !strings.Contains(duck, "o") || !strings.Contains(duck, "good") {
		t.Errorf("Happy duck should contain expected pattern")
	}
}

func TestGetAngryDuck(t *testing.T) {
	duck := GetAngryDuck()
	if duck == "" {
		t.Errorf("GetAngryDuck() should return non-empty string")
	}
	// Should contain QUACK
	if !strings.Contains(duck, "QUACK") {
		t.Errorf("Angry duck should contain QUACK")
	}
}

func TestGetBanner(t *testing.T) {
	banner := GetBanner()
	if banner == "" {
		t.Errorf("GetBanner() should return non-empty string")
	}
	// Banner should contain some reference to environment or duck
	if !strings.Contains(banner, "Environment") && !strings.Contains(banner, "🦆") {
		t.Errorf("Banner should contain environment or duck reference")
	}
}

func TestGetSyncMessage(t *testing.T) {
	msg := GetSyncMessage()
	if msg == "" {
		t.Errorf("GetSyncMessage() should return non-empty string")
	}
}

func TestDucksAreDifferent(t *testing.T) {
	happy := GetHappyDuck()
	angry := GetAngryDuck()

	// Happy and angry ducks should be different
	if happy == angry {
		t.Errorf("Happy and angry ducks should be different")
	}
}
