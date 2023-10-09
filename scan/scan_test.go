package scan

import (
	"os/user"
	"path/filepath"
	"testing"
)

func TestScanGitFolders(t *testing.T) {
	usr, err := user.Current()
    if err != nil {
        t.Fatal(err)
	}
	testFolder := usr.HomeDir + "/Coding" 

	expectedFolders := []string{
		filepath.Join(testFolder, ".git"),
	}

	folders := RecursiveScanFolder(testFolder)

	for _, expected := range expectedFolders {
		found := false
		for _, folder := range folders {
			if folder == expected {
				found = true
				break
			}
		}
		if !found {
			t.Log("not contains .git folder")
		}
	}
}

func BenchmarkRecursiveScanFolder(b *testing.B) {
    usr, err := user.Current()
    if err != nil {
        b.Fatal(err)
	}

	b.ResetTimer()
	testFolder := usr.HomeDir + "/Coding"
	for j := 0; j < b.N; j++ {
		_ = RecursiveScanFolder(testFolder)
	}
}