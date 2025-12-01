package config

import (
	"os"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	if _, err := New("./example.yaml"); err != nil {
		t.Errorf("Unable to create a new Config: %v", err)
	}
}

func TestDirsIsCreated(t *testing.T) {
	if _, err := os.Stat("./tmp"); os.IsNotExist(err) {
		t.Errorf("Tmp dir is not created: %v", err)
	}
	if _, err := os.Stat("./tmp/download"); os.IsNotExist(err) {
		t.Errorf("Download dir is not created: %v", err)
	}

	if _, err := os.Stat("./tmp/download/cat1"); os.IsNotExist(err) {
		t.Errorf("Cat1 dir is not created: %v", err)
	}

	if _, err := os.Stat("./tmp/download/cat2"); os.IsNotExist(err) {
		t.Errorf("Cat2 dir is not created: %v", err)
	}

	if _, err := os.Stat("./tmp/images"); os.IsNotExist(err) {
		t.Errorf("Images dir is not created: %v", err)
	}
}

func TestRmTmpDir(t *testing.T) {
	if err := os.RemoveAll("./tmp"); err != nil {
		t.Errorf("Tmp dir is not removed: %v", err)
	}
}
