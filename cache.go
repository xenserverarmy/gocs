package gocs

import (
	"bitbucket.org/kardianos/osext"
	"encoding/gob"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

var cacheFile string

func setCacheFile() error {
	if cacheFile == "" {
		exe, err := osext.Executable()
		if err != nil {
			return err
		}
		strings.TrimSuffix(exe, path.Ext(exe))
		cacheFile = exe + ".cache"
	}
	return nil
}

// Updates (or creates) a cache containing the API's structure
func cacheApiCommands(cs *CloudStackClient, cacheExpirationInDays int) (err error) {
	if cacheExpirationInDays > 0 {
		if err := setCacheFile(); err != nil {
			return err
		}
		beforeDate := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-cacheExpirationInDays, time.Now().Hour(), time.Now().Minute(), time.Now().Second(), time.Now().Nanosecond(), time.Local)
		if fileInfo, err := os.Stat(cacheFile); err != nil || fileInfo.ModTime().Before(beforeDate) {
			if err := updateApiCommandFile(cs); err != nil {
				return err
			}
		}
		cs.apiCommands, err = readApiCommandFile(cs)
	} else {
		cs.apiCommands, err = requestApiCommands(cs)
	}
	return
}

func readApiCommandFile(cs *CloudStackClient) (commands, error) {
	reader, err := os.Open(cacheFile)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	decoder := gob.NewDecoder(reader)
	var apiCommands commands
	if err := decoder.Decode(&apiCommands); err != nil {
		return nil, err
	}
	return apiCommands, nil
}

// Calls newRequest() directly as there is no cache yet, so it's not
// possible to check and verify this call which is done by the exported
// request functions
func requestApiCommands(cs *CloudStackClient) (commands, error) {
	rawJSON, err := newRequest(cs, &command{Name: "listApis"}, url.Values{})
	if err != nil {
		return nil, err
	}
	return unmarshalApiCommands(rawJSON)
}

func updateApiCommandFile(cs *CloudStackClient) error {
	result, err := requestApiCommands(cs)
	if err != nil {
		return err
	}

	writer, err := os.Create(cacheFile)
	if err != nil {
		return err
	}
	defer writer.Close()

	encoder := gob.NewEncoder(writer)
	return encoder.Encode(result)
}
