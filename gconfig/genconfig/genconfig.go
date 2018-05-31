package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type Config struct {
	Output string
	Tables [][]interface{}
}

var (
	config Config

	header = `package gconfig //自动生成，不要手动修改

	import (
		"fmt"
		"io/ioutil"
	
		"github.com/json-iterator/go"
		. "github.com/sencydai/gameworld/typedefine"
	)
	
	var (
		json = jsoniter.ConfigCompatibleWithStandardLibrary
	)
	
	func loadConfig(path, name string, v interface{}) {
		if data, err := ioutil.ReadFile(fmt.Sprintf("%s/%s.json", path, name)); err != nil {
			panic(err)
		} else if !json.Valid(data) {
			panic(fmt.Errorf("parse config %s failed", name))
		} else if err = json.Unmarshal(data, v); err != nil {
			panic(err)
		}
	}
	`
)

func getMapType(keys int) string {
	if keys == 0 {
		return ""
	}

	return "map[int]" + getMapType(keys-1)
}

func main() {
	if buff, err := ioutil.ReadFile("config.json"); err != nil {
		fmt.Printf("load config file error: %s\n", err.Error())
		return
	} else if err = json.Unmarshal(buff, &config); err != nil {
		fmt.Printf("parse config file error: %s\n", err.Error())
		return
	}

	varDefines := make([]string, 0)
	loadConfigs := make([]string, 0)

	for _, tables := range config.Tables {
		name := tables[0].(string)
		keys := int(tables[1].(float64))
		varTypes := getMapType(keys) + name

		varDefines = append(varDefines, fmt.Sprintf("G%s %s", name, varTypes))

		configs := make([]string, 3)
		if keys == 0 {
			configs[0] = fmt.Sprintf("g%s := %s{}", name, varTypes)
		} else {
			configs[0] = fmt.Sprintf("g%s := make(%s)", name, varTypes)
		}
		configs[1] = fmt.Sprintf("loadConfig(path,\"%s\",&g%s)", name, name)
		configs[2] = fmt.Sprintf("G%s = g%s", name, name)

		loadConfigs = append(loadConfigs, strings.Join(configs, "\n")+"\n")
	}

	file, err := os.Create(fmt.Sprintf("%s/gconfig.go", config.Output))
	if err != nil {
		fmt.Println(err)
		return
	}
	file.WriteString(header)
	file.WriteString(fmt.Sprintf("var (\n%s\n)\n\n", strings.Join(varDefines, "\n")))
	file.WriteString(fmt.Sprintf("func LoadConfigs(path string) {\n%s}", strings.Join(loadConfigs, "\n")))
	file.Sync()
	file.Close()

	cmd := exec.Command("gofmt", "-w", fmt.Sprintf("%s/gconfig.go", config.Output))
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
}
