package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	wpDir string
	siteDir string
	autoCreateMissingDir bool
)

type WpFile struct {
	Name string
	FileInfo os.FileInfo
}

func isValidWordpressDir(dir string) bool {
	// Here we just check to see if we have a wp-activate.php file
	validDir := true
	wpActivatePath := path.Join(dir, "wp-activate.php")
	if _, err := os.Stat(wpActivatePath); os.IsNotExist(err) {
		validDir = false
	}
	return validDir
}

func init() {
	flag.StringVar(&wpDir, "wp-dir", "", "The Wordpress source directory")
	flag.StringVar(&siteDir, "site-dir", "", "The directory to upgrade")
	flag.BoolVar(&autoCreateMissingDir, "create-missing-dir", true, "Automatically create directories that are missing")
	flag.Parse()

	hasErrs := false
	if wpDir == "" {
		fmt.Println("The base Wordpress source directory must be specified")
		hasErrs = true
	} else if !isValidWordpressDir(wpDir) {
		fmt.Printf("The base Wordpress source directory %s does not appear to be valid.\n", wpDir)
		hasErrs = true
	}

	if siteDir == "" {
		fmt.Println("The site Wordpress installation directory must be specified")
		hasErrs = true
	} else if !isValidWordpressDir(siteDir) {
		fmt.Printf("The base Wordpress source directory %s does not appear to be valid.\n", siteDir)
		hasErrs = true
	}

	if hasErrs {
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	// Now traverse through our list of files from the source directory,
	wpFiles := make([]WpFile, 0)
	pathMap := make(map[string]WpFile)

	filepath.Walk(wpDir, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}

		// strip the wpDir prefix from the path
		if len(filePath) > len(wpDir) {
			abbrPath := filePath[len(wpDir)+1:]

			if abbrPath[0:1] == "." {
				return nil // ignore files that start with a dot
			}

			wpFile := WpFile{
				Name:     abbrPath,
				FileInfo: fileInfo,
			}
			wpFiles = append(wpFiles, wpFile)
			pathMap[abbrPath] = wpFile
		}

		return nil
	})

	// now check to see if all the directories exist in the dest path
	pathErrs := false
	lastMissingPath := ""
	for _, wpFile := range wpFiles {
		if wpFile.FileInfo.IsDir() {
			if lastMissingPath == "" || !strings.HasPrefix(wpFile.Name, lastMissingPath) {
				destDir := path.Join(siteDir, wpFile.Name)
				if _, err := os.Stat(destDir); os.IsNotExist(err) {
					fmt.Printf("Destination directory %s does not exist!\n", destDir)
					if autoCreateMissingDir {
						err = os.Mkdir(destDir, 0755)
						if err != nil {
							panic(err)
						}
					} else {
						lastMissingPath = wpFile.Name
						pathErrs = true
					}
				}
			}
		}
	}

	if pathErrs {
		fmt.Println("Fix missing directories and then re-run script!")
		os.Exit(2)
	}

	for _, wpFile := range wpFiles {
		if !wpFile.FileInfo.IsDir() {
			sourceFile := path.Join(wpDir, wpFile.Name)
			destFile := path.Join(siteDir, wpFile.Name)

			input, err := ioutil.ReadFile(sourceFile)
			if err != nil {
				panic(err)
			}

			fileMode := os.FileMode(0644)
			fileInfo, err := os.Stat(destFile)
			if os.IsExist(err) {
				fileMode = fileInfo.Mode()
			}

			err = ioutil.WriteFile(destFile, input, fileMode)
			if err != nil {
				panic(err)
			}
		}

	}
}
