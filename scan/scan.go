package scan

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
)

// scan scans a new folder for Git repositories
func Scan(folder string) {
	fmt.Printf("Found folders:\n\n")
	repositories := RecursiveScanFolder(folder)
	filePath := GetDotFilePath()
	addNewSliceElementsToFile(filePath, repositories)
	fmt.Printf("\n\nSuccessfully added\n\n")
}

// addNewSliceElementsToFile given a slice of strings representing paths, stores them
// to the filesystem
func addNewSliceElementsToFile(filePath string, newRepos []string) {
	existingRepos := ParseFileLinesToSlice(filePath)
	repos := joinSlices(newRepos, existingRepos)
	DumpStringsSliceToFile(repos, filePath)
}

// dumpStringsSliceToFile writes content to the file in path `filePath` (overwriting existing content)
func DumpStringsSliceToFile(repos []string, filepath string) {
	content := strings.Join(repos, "\n")
	os.WriteFile(filepath, []byte(content), 0755)
}

// joinSlices adds the element of the `new` slice
// into the `existing` slice, only if not already there
func joinSlices(new []string, existing []string) []string {
	for _, i := range new {
		if !SliceContains(existing, i) {
			existing = append(existing, i)
		}
	}
	return existing
}

// sliceContains returns true if `slice` contains `value`
func SliceContains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// parseFileLinesToSlice given a file path string, gets the content
// of each line and parses it to a slice of strings.
func ParseFileLinesToSlice(filepath string) []string {
	f := OpenFile(filepath)
	defer f.Close()
	fmt.Println(filepath)
	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		if err != io.EOF {
			panic(err)
		}
	}

	return lines
}

// openFile opens the file located at `filePath`. Creates it if not existing.
func OpenFile(filepath string) *os.File {
	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_RDWR, 0755)
	if err != nil {
		if os.IsNotExist(err) {
			// file does not exist
			_, err = os.Create(filepath)
			if err != nil {
				panic(err)
			} else {
				// other error
				panic(err)
			}
		}
	}

	return f
}

// getDotFilePath returns the dot file for the repos list.
// Creates it and the enclosing folder if it does not exist.
func GetDotFilePath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	dotFile := usr.HomeDir + "/.gitlocalstats"

	// Check if the file exists, and create it if it doesn't.
	if _, err := os.Stat(dotFile); os.IsNotExist(err) {
		if _, err := os.Create(dotFile); err != nil {
			log.Fatal(err)
		}
	}

	return dotFile
}

// RecursiveScanFolder starts the recursive search of git repositories
// living in the `folder` subtree
func RecursiveScanFolder(folder string) []string {
	folders := make([]string, 0)
	ch := make(chan string, 1000) // Buffered channel to receive paths
	var wg sync.WaitGroup

	wg.Add(1)
	go scanGitFolders(&folders, folder, ch, &wg)

	go func() {
		wg.Wait()
		close(ch)
	}()

	for path := range ch {
		folders = append(folders, path)
	}

	return folders
}

// scanGitFolders returns a list of subfolders of `folder` ending with `.git`.
// Returns the base folder of the repo, the .git folder parent.
// Recursively searches in the subfolders by passing an existing `folders` slice.
func scanGitFolders(folders *[]string, folder string, ch chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	folder = strings.TrimSuffix(folder, "/")

	files, err := os.ReadDir(folder)
	if err != nil {
		log.Fatal(err)
	}

	var path string

	for _, file := range files {
		if file.IsDir() {
			path = filepath.Join(folder, file.Name())

			// Folder is a git repo
			if file.Name() == ".git" {
				path = strings.TrimSuffix(path, "/.git")
				ch <- path
			} else if file.Name() != "vendor" && file.Name() != "node_modules" {
				// Recursively scan file systems
				wg.Add(1)
				go scanGitFolders(folders, path, ch, wg)
			}
		}
	}
}
