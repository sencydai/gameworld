package gconfig

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const (
	fileName = "randomnames.txt"
)

var (
	randomUnUseNames map[string]bool
	randomUsedNames  = make(map[string]bool)
)

func LoadRandomNames(path string) {
	f, err := os.Open(fmt.Sprintf("%s/%s", path, fileName))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	names := make(map[string]bool)
	reader := bufio.NewReader(f)
	for {
		line, _, _ := reader.ReadLine()
		if len(line) == 0 {
			break
		}
		names[strings.Trim(string(line), " ")] = true
	}
	for name := range randomUsedNames {
		if _, ok := names[name]; !ok {
			delete(randomUsedNames, name)
		} else {
			delete(names, name)
		}
	}
	randomUnUseNames = names
}

func GetRandomName() string {
	for name := range randomUnUseNames {
		return name
	}

	return ""
}

func UseRandomName(name string) {
	if _, ok := randomUnUseNames[name]; ok {
		delete(randomUnUseNames, name)
		randomUsedNames[name] = true
	}
}

func FreeRandomName(name string) {
	if _, ok := randomUsedNames[name]; ok {
		delete(randomUsedNames, name)
		randomUnUseNames[name] = true
	}
}
