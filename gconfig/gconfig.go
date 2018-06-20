package gconfig //自动生成，不要手动修改

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

var (
	GLordBaseConfig        LordBaseConfig
	GLordConfig            map[int]map[int]LordConfig
	GLordLevelConfig       map[int]LordLevelConfig
	GLordEquipConfig       map[int]LordEquipConfig
	GLordEquipStrengConfig map[int]map[int]LordEquipStrengConfig
	GItemConfig            map[int]ItemConfig
	GItemGroupConfig       map[int]map[int]ItemGroupConfig
	GVirtualCurrencyConfig map[int]VirtualCurrencyConfig
)

func LoadConfigs(path string) {
	gLordBaseConfig := LordBaseConfig{}
	loadConfig(path, "LordBaseConfig", &gLordBaseConfig)
	GLordBaseConfig = gLordBaseConfig

	gLordConfig := make(map[int]map[int]LordConfig)
	loadConfig(path, "LordConfig", &gLordConfig)
	GLordConfig = gLordConfig

	gLordLevelConfig := make(map[int]LordLevelConfig)
	loadConfig(path, "LordLevelConfig", &gLordLevelConfig)
	GLordLevelConfig = gLordLevelConfig

	gLordEquipConfig := make(map[int]LordEquipConfig)
	loadConfig(path, "LordEquipConfig", &gLordEquipConfig)
	GLordEquipConfig = gLordEquipConfig

	gLordEquipStrengConfig := make(map[int]map[int]LordEquipStrengConfig)
	loadConfig(path, "LordEquipStrengConfig", &gLordEquipStrengConfig)
	GLordEquipStrengConfig = gLordEquipStrengConfig

	gItemConfig := make(map[int]ItemConfig)
	loadConfig(path, "ItemConfig", &gItemConfig)
	GItemConfig = gItemConfig

	gItemGroupConfig := make(map[int]map[int]ItemGroupConfig)
	loadConfig(path, "ItemGroupConfig", &gItemGroupConfig)
	GItemGroupConfig = gItemGroupConfig

	gVirtualCurrencyConfig := make(map[int]VirtualCurrencyConfig)
	loadConfig(path, "VirtualCurrencyConfig", &gVirtualCurrencyConfig)
	GVirtualCurrencyConfig = gVirtualCurrencyConfig
}
