package builder

import "testing"

func TestLocalConfigRoundTripWindowsPaths(t *testing.T) {
	root := t.TempDir()
	want := LocalConfig{
		GameRoot: `D:\Games\Evenicle`,
		AliceBin: `C:\Tools\AliceTools\alice.exe`,
	}
	if err := SaveLocalConfig(root, want); err != nil {
		t.Fatal(err)
	}
	got, err := LoadLocalConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("round-trip mismatch: got %#v want %#v", got, want)
	}
}
