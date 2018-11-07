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
	sourceDir            string
	destDir              string
	minPctFilesExist     float64
	autoCreateMissingDir bool
)

type WpFile struct {
	Name     string
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
	flag.StringVar(&sourceDir, "src-dir", "", "The source directory to copy")
	flag.StringVar(&destDir, "dest-dir", "", "The directory to upgrade")
	flag.Float64Var(&minPctFilesExist, "min-pct", 0.60, "The minimum percentage of files that need to exist.")
	flag.BoolVar(&autoCreateMissingDir, "create-missing-dir", true, "Automatically create directories that are missing")
	flag.Parse()

	hasErrs := false
	if sourceDir == "" {
		fmt.Println("The base source directory must be specified")
		hasErrs = true
	} else {
		sourceDir = path.Clean(sourceDir)
		if !isValidWordpressDir(sourceDir) {
			fmt.Printf("The base Wordpress source directory %s does not appear to be valid.\n", sourceDir)
			hasErrs = true
		}
	}

	if destDir == "" {
		fmt.Println("The site Wordpress installation directory must be specified")
		hasErrs = true
	} else {
		destDir = path.Clean(destDir)
		if !isValidWordpressDir(destDir) {
			fmt.Printf("The base Wordpress source directory %s does not appear to be valid.\n", destDir)
			hasErrs = true
		}
	}

	if hasErrs {
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	// Now traverse through our list of files from the source directory,
	srcFiles := make([]WpFile, 0)
	pathMap := make(map[string]WpFile)

	filepath.Walk(sourceDir, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}

		// strip the sourceDir prefix from the path
		if len(filePath) > len(sourceDir) {
			abbrPath := filePath[len(sourceDir)+1:]

			if abbrPath[0:1] == "." {
				return nil // ignore files that start with a dot
			}

			wpFile := WpFile{
				Name:     abbrPath,
				FileInfo: fileInfo,
			}
			srcFiles = append(srcFiles, wpFile)
			pathMap[abbrPath] = wpFile
		}

		return nil
	})

	// now check to see if all the directories exist in the dest path
	pathErrs := false
	lastMissingPath := ""
	numFilesExisting := 0.0
	numFilesInTotal := 0.0
	for _, srcFile := range srcFiles {
		if srcFile.FileInfo.IsDir() {
			if lastMissingPath == "" || !strings.HasPrefix(srcFile.Name, lastMissingPath) {
				destDir := path.Join(destDir, srcFile.Name)
				if _, err := os.Stat(destDir); os.IsNotExist(err) {
					fmt.Printf("Destination directory %s does not exist!\n", destDir)
					if autoCreateMissingDir {
						err = os.Mkdir(destDir, 0755)
						if err != nil {
							panic(err)
						}
					} else {
						lastMissingPath = srcFile.Name
						pathErrs = true
					}
				}
			}
		} else {
			numFilesInTotal += 1.0
			destDir := path.Join(destDir, srcFile.Name)
			if _, err := os.Stat(destDir); os.IsExist(err) {
				numFilesExisting += 1.0
			}
		}
	}

	if pathErrs {
		fmt.Println("Fix missing directories and then re-run script!")
		os.Exit(2)
	}

	if numFilesInTotal == 0.0 {
		fmt.Println("Number of files in source directory is zero")
		os.Exit(3)
	}

	pctFilesExist := numFilesExisting / numFilesInTotal
	if pctFilesExist < minPctFilesExist {
		fmt.Printf("Percentage of files that exist in dest directory %.2f%% is less than minimum %.2f%%\n", pctFilesExist*100.0, minPctFilesExist*100.0)
		os.Exit(4)
	}

	for _, wpFile := range srcFiles {
		if !wpFile.FileInfo.IsDir() {
			sourceFile := path.Join(sourceDir, wpFile.Name)
			destFile := path.Join(destDir, wpFile.Name)

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
