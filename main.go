package main

import (
	"fmt"
	"github.com/nyulibraries/go-ead-indexer/pkg/ead"
	eadutil "github.com/nyulibraries/go-ead-indexer/pkg/ead/eadutil"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const xmlFileSuffix = "-add.xml"

var eadDirPath string
var httpRequestXMLBodiesPrettifiedDirPath string
var httpRequestXMLBodiesRawDirPath string
var rootPath string

// We need to get the absolute path to this package in order to get the absolute
// path to the tmp/ directory.  We don't want the wrong directories clobbered by
// the output if this script is run from somewhere outside of this directory.
func init() {
	// The `filename` string is the absolute path to this source file, which should
	// be located at the root of the package directory.
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Panic("ERROR: `runtime.Caller(0)` failed")
	}

	rootPath = filepath.Dir(filename)
}

func abortBadUsage(err error) {
	if err != nil {
		log.Println(err.Error())
	}
	usage()
	os.Exit(1)
}

func clean() error {
	err := os.RemoveAll(httpRequestXMLBodiesPrettifiedDirPath)
	if err != nil {
		return err
	}
	err = os.MkdirAll(httpRequestXMLBodiesPrettifiedDirPath, 0700)
	if err != nil {
		return err
	}

	err = os.RemoveAll(httpRequestXMLBodiesRawDirPath)
	if err != nil {
		return err
	}
	err = os.MkdirAll(httpRequestXMLBodiesRawDirPath, 0700)
	if err != nil {
		return err
	}

	return nil
}

func getEADFilePath(testEAD string) string {
	return filepath.Join(eadDirPath, testEAD+".xml")
}

func getEADValue(testEAD string) (string, error) {
	return getFileContents(getEADFilePath(testEAD))
}

func getFileContents(filename string) (string, error) {
	bytes, err := os.ReadFile(filename)

	if err != nil {
		return filename, err
	}

	return string(bytes), nil
}

func getFullyQualifiedEADs() []string {
	fullyQualifiedEADs := []string{}

	err := filepath.WalkDir(eadDirPath, func(path string, dirEntry fs.DirEntry, err error) error {
		if !dirEntry.IsDir() && filepath.Ext(path) == ".xml" {
			repositoryCode := filepath.Base(filepath.Dir(path))
			eadID := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
			fullyQualifiedEADs = append(fullyQualifiedEADs, fmt.Sprintf("%s/%s", repositoryCode, eadID))
		}
		return nil
	})
	if err != nil {
		log.Panic(fmt.Sprintf(`getFullyQualifiedEADs() failed: %s`, err))

	}

	return fullyQualifiedEADs
}

func isDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Panic(fmt.Sprintf(`isDirectory("%s") failed with error: %s`, path, err))
	}

	return fileInfo.IsDir()
}

func parseEADID(testEAD string) string {
	return filepath.Base(testEAD)
}

func parseRepositoryCode(testEAD string) string {
	return filepath.Dir(testEAD)
}

func setDirectoryPaths() {
	args := os.Args

	if len(args) < 2 {
		abortBadUsage(fmt.Errorf("Wrong number of args"))
	}

	// Declare `err` instead of doing `eadDirPath, err :=`, which shadows package
	// level var `eadDirPath`.
	var err error
	eadDirPath, err = filepath.Abs(args[1])
	// Very basic validation of directory:
	// - No error when resolving to absolute path
	// - Is a directory and not a symlink -- because `filepath.WalkDir` does not
	//   work on symlinks.
	// - Directory name matches repo name.  This is not strictly necessary, because
	//   we should be able to rename the directory if we want, but this is just
	//   a one-off test (for now -- maybe we'll make a permanent one later).
	//   This also has the benefit of preventing possible accidental use of the
	//   FABified repo, which has a different name.
	if err != nil || !isDirectory(eadDirPath) || !strings.HasSuffix(eadDirPath, "findingaids_eads_v2") {
		abortBadUsage(fmt.Errorf(`Path "%s" is not a valid findingaids_eads_v2 repo path`, args[1]))
	}

	httpRequestXMLBodiesRawDirPath = filepath.Join(rootPath, "http-requests-xml", "raw")
	httpRequestXMLBodiesPrettifiedDirPath = filepath.Join(rootPath, "http-requests-xml", "prettified")
}

func usage() {
	log.Println("usage: go run main.go [path to findingaids_eads_v2]")
}

func writeSolrAddXMLFiles(fullyQualifiedEAD string, fileID string,
	rawXML string, prettifiedXML string) error {

	prettifiedFile := xmlFilePrettified(fullyQualifiedEAD, fileID)
	err := os.MkdirAll(filepath.Dir(prettifiedFile), 0755)
	if err != nil {
		return err
	}
	err = os.WriteFile(prettifiedFile, []byte(prettifiedXML), 0644)
	if err != nil {
		return err
	}

	rawFile := xmlFileRaw(fullyQualifiedEAD, fileID)
	err = os.MkdirAll(filepath.Dir(rawFile), 0755)
	if err != nil {
		return err
	}

	return os.WriteFile(rawFile, []byte(rawXML), 0644)
}

func xmlFilePrettified(fullyQualifiedEAD string, fileID string) string {
	return filepath.Join(httpRequestXMLBodiesPrettifiedDirPath, fullyQualifiedEAD, fileID+xmlFileSuffix)
}

func xmlFileRaw(fullyQualifiedEAD string, fileID string) string {
	return filepath.Join(httpRequestXMLBodiesRawDirPath, fullyQualifiedEAD, fileID+xmlFileSuffix)
}

func main() {
	setDirectoryPaths()

	err := clean()
	if err != nil {
		log.Panic("clean() error: " + err.Error())
	}

	fullyQualifiedEADs := getFullyQualifiedEADs()

	for _, fullyQualifiedEAD := range fullyQualifiedEADs {
		fmt.Printf("[ %s ] Writing files for %s\n", time.Now().Format("2006-01-02 15:04:05"), fullyQualifiedEAD)
		eadXML, err := getEADValue(fullyQualifiedEAD)
		if err != nil {
			log.Println(fmt.Sprintf(`getEADValue("%s") failed: %s`, fullyQualifiedEAD, err))
		}

		repositoryCode := parseRepositoryCode(fullyQualifiedEAD)
		eadToProcess, err := ead.New(repositoryCode, eadXML)
		if err != nil {
			log.Println(fmt.Sprintf(`ead.New("%s", [EADXML for %s ]) failed: %s`, repositoryCode, fullyQualifiedEAD, err))
		}

		eadID := parseEADID(fullyQualifiedEAD)
		rawXML := eadToProcess.CollectionDoc.SolrAddMessage.String()
		prettifiedXML := eadutil.PrettifySolrAddMessageXML(rawXML)
		err = writeSolrAddXMLFiles(fullyQualifiedEAD, eadID, rawXML, prettifiedXML)
		if err != nil {
			log.Println(err.Error())
		}

		if eadToProcess.Components == nil {
			fmt.Println(fullyQualifiedEAD + " has no components.  Skipping component XML file writing")

			continue
		}

		componentIDs := []string{}
		for _, component := range *eadToProcess.Components {
			componentIDs = append(componentIDs, component.ID)
			rawXML := component.SolrAddMessage.String()
			prettifiedXML := eadutil.PrettifySolrAddMessageXML(rawXML)
			err = writeSolrAddXMLFiles(fullyQualifiedEAD, component.ID, rawXML, prettifiedXML)
			if err != nil {
				log.Println(err.Error())
			}
		}
	}
}
