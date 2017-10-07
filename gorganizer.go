package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/disiqueira/gotree"
	"gopkg.in/ini.v1"
)

const (
	configFile = ".gorganizer-{lang}.ini"
)

var cfg *ini.File
var cfgFile string
var language string

func addToTree(folder, file string, tree *gotree.GTStructure) {

	newFile := gotree.GTStructure{Name: file}

	// append to parent, if exists
	for i, item := range tree.Items {
		if item.Name == folder {
			item.Items = append(item.Items, newFile)
			tree.Items[i] = item
			return
		}
	}

	// create parent if missing
	newFolder := gotree.GTStructure{Name: folder}
	newFolder.Items = append(newFolder.Items, newFile)
	tree.Items = append(tree.Items, newFolder)
}

type excludeListType []string

var excludeList excludeListType

func (e excludeListType) checkExclude(ext string) bool {
	if ext == "" {
		return false
	}
	for _, item := range e {
		if item == ext {
			return true
		}
	}
	return false
}

func main() {
	outputFolder := flag.String("output", ".", "Main directory to put organized folders")
	inputFolder := flag.String("directory", ".", "The directory whose files to classify")

	newRule := flag.String("newrule", "", "Insert a new rule. Format ext:folder Example: mp3:Music")
	delRule := flag.String("delrule", "", "Delete a rule. Format ext Example: mp3")

	printRules := flag.Bool("allrules", false, "Print all rules")

	preview := flag.Bool("preview", false, "Only preview, do not move files")
	recursive := flag.Bool("recursive", false, "Search over all directories.")

	excludeExtentions := flag.String("exclude", "", "Exclude files will ignore files for organizer. Format pdf,odt")

	flag.StringVar(&language, "language", "en", "Specify language: en|tr|pt")

	flag.Parse()

	initDb()

	defer closeDb()

	excludeList = strings.Split(*excludeExtentions, ",")

	if len(*newRule) > 0 {
		fmt.Println("Creating new rule")
		err := insertRule(*newRule)
		if err != nil {
			fmt.Println(err)
			return
		}
		showRules()
		return
	}

	if len(*delRule) > 0 {
		fmt.Println("Deleting rule")
		deleteRule(*delRule)
		showRules()
		return
	}

	if *printRules {
		showRules()
		return
	}

	fmt.Println("GOrganizing your Files")

	var tree gotree.GTStructure
	tree.Name = "Files"

	scanDirectory(*inputFolder, *outputFolder, &tree, *preview, *recursive)

	gotree.PrintTree(tree)

	fmt.Println("All files have been GOrganized!")
}

func scanDirectory(inputFolder, outputFolder string, tree *gotree.GTStructure, preview, recursive bool) {
	files, _ := ioutil.ReadDir(inputFolder)
	for _, f := range files {
		if f.IsDir() && recursive {
			scanDirectory(filepath.Join(inputFolder, f.Name()), outputFolder, tree, preview, recursive)
		}

		file := filepath.Join(inputFolder, f.Name())
		ext := strings.TrimPrefix(path.Ext(file), ".")

		if excludeList.checkExclude(ext) {
			addToTree("Excluded Files", f.Name(), tree)
			continue
		}

		newFolder := iniGet(ext)

		if len(newFolder) > 0 {

			folder := filepath.Join(outputFolder, newFolder)
			newFile := filepath.Join(folder, f.Name())

			if !preview {
				_ = os.Mkdir(folder, os.ModePerm)
				os.Rename(file, newFile)
			}
		} else {
			newFolder = "Unknown extension (will not be moved)"
		}

		addToTree(newFolder, f.Name(), tree)
	}
}
