package gconfig //自动生成，不要手动修改

import (
	"fmt"
	"io/ioutil"

	"github.com/json-iterator/go"
	t "github.com/sencydai/gameworld/typedefine"
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
	GLordBaseConfig            *t.LordBaseConfig
	GLordConfig                map[int]map[int]*t.LordConfig
	GLordLevelConfig           map[int]*t.LordLevelConfig
	GLordEquipConfig           map[int]*t.LordEquipConfig
	GLordEquipStrengConfig     map[int]map[int]*t.LordEquipStrengConfig
	GLordHeadConfig            map[int]*t.LordHeadConfig
	GLordFrameConfig           map[int]*t.LordFrameConfig
	GLordChatConfig            map[int]*t.LordChatConfig
	GChangeJobConfig           map[int]*t.ChangeJobConfig
	GItemConfig                map[int]*t.ItemConfig
	GItemGroupConfig           map[int]map[int]*t.ItemGroupConfig
	GVirtualCurrencyConfig     map[int]*t.VirtualCurrencyConfig
	GTalentConfig              map[int]*t.TalentConfig
	GTalentLevelConfig         map[int]map[int]*t.TalentLevelConfig
	GLordSkillConfig           map[int]map[int]*t.LordSkillConfig
	GLordSkillLevelConfig      map[int]map[int]map[int]*t.LordSkillLevelConfig
	GLordSkillCostConfig       map[int]*t.LordSkillCostConfig
	GGuardLevelConfig          map[int]*t.GuardLevelConfig
	GGuardModelConfig          map[int]*t.GuardModelConfig
	GHeroBaseConfig            *t.HeroBaseConfig
	GHeroConfig                map[int]*t.HeroConfig
	GHeroRawConfig             map[int]*t.HeroRawConfig
	GHeroRarityConfig          map[int]*t.HeroRarityConfig
	GHeroLevelConfig           map[int]map[int]*t.HeroLevelConfig
	GHeroStageConfig           map[int]*t.HeroStageConfig
	GHeroStageTalentConfig     map[int]map[int]*t.HeroStageTalentConfig
	GHeroChangeJobConfig       map[int]map[int]*t.HeroChangeJobConfig
	GHeroChangeJobCostConfig   map[int]map[int]*t.HeroChangeJobCostConfig
	GHeroFetterConfig          map[int]map[int]*t.HeroFetterConfig
	GHeroOpenConfig            map[int]map[int]*t.HeroOpenConfig
	GMonsterConfig             map[int]*t.MonsterConfig
	GMonsterTemplateConfig     map[int]map[int]*t.MonsterTemplateConfig
	GMonsterLordConfig         map[int]*t.MonsterLordConfig
	GMonsterLordTemplateConfig map[int]map[int]*t.MonsterLordTemplateConfig
	GMainFubenBaseConfig       *t.MainFubenBaseConfig
	GMainFubenConfig           map[int]*t.MainFubenConfig
	GBuffConfig                map[int]*t.BuffConfig
	GBuffBaseConfig            map[int]*t.BuffBaseConfig
	GSkillConfig               map[int]*t.SkillConfig
	GSkillEffectConfig         map[int]map[int]*t.SkillEffectConfig
	GRaceConfig                map[int]*t.RaceConfig
	GEquipBaseConfig           *t.EquipBaseConfig
	GEquipConfig               map[int]*t.EquipConfig
	GEquipPosConfig            map[int]*t.EquipPosConfig
	GEquipRarityConfig         map[int]*t.EquipRarityConfig
	GEquipStrengConfig         map[int]map[int]*t.EquipStrengConfig
	GEquipSuitConfig           map[int]*t.EquipSuitConfig
	GEquipStrengMasterConfig   map[int]*t.EquipStrengMasterConfig
	GArtifactBaseConfig        *t.ArtifactBaseConfig
	GArtifactConfig            map[int]*t.ArtifactConfig
	GArtifactRarityConfig      map[int]map[int]*t.ArtifactRarityConfig
	GArtifactStrengConfig      map[int]map[int]*t.ArtifactStrengConfig
	GBuildingConfig            map[int]*t.BuildingConfig
	GBuildingLevelConfig       map[int]map[int]*t.BuildingLevelConfig
	GChatBaseConfig            *t.ChatBaseConfig
	GSystemNoticeConfig        map[int]map[int]map[int]*t.SystemNoticeConfig
	GSystemTimingNoticeConfig  map[int]*t.SystemTimingNoticeConfig
	GSystemConfig              *t.SystemConfig
	GSystemOpenConfig          map[int]*t.SystemOpenConfig
	GFightSkipConfig           map[int]*t.FightSkipConfig
	GWorldBossBaseConfig       *t.WorldBossBaseConfig
	GWorldBossConfig           map[int]*t.WorldBossConfig
	GWorldBossRankConfig       map[int]map[int]*t.WorldBossRankConfig
)

func LoadConfigs(path string) {
	gLordBaseConfig := &t.LordBaseConfig{}
	loadConfig(path, "LordBaseConfig", gLordBaseConfig)
	GLordBaseConfig = gLordBaseConfig

	gLordConfig := make(map[int]map[int]*t.LordConfig)
	loadConfig(path, "LordConfig", &gLordConfig)
	GLordConfig = gLordConfig

	gLordLevelConfig := make(map[int]*t.LordLevelConfig)
	loadConfig(path, "LordLevelConfig", &gLordLevelConfig)
	GLordLevelConfig = gLordLevelConfig

	gLordEquipConfig := make(map[int]*t.LordEquipConfig)
	loadConfig(path, "LordEquipConfig", &gLordEquipConfig)
	GLordEquipConfig = gLordEquipConfig

	gLordEquipStrengConfig := make(map[int]map[int]*t.LordEquipStrengConfig)
	loadConfig(path, "LordEquipStrengConfig", &gLordEquipStrengConfig)
	GLordEquipStrengConfig = gLordEquipStrengConfig

	gLordHeadConfig := make(map[int]*t.LordHeadConfig)
	loadConfig(path, "LordHeadConfig", &gLordHeadConfig)
	GLordHeadConfig = gLordHeadConfig

	gLordFrameConfig := make(map[int]*t.LordFrameConfig)
	loadConfig(path, "LordFrameConfig", &gLordFrameConfig)
	GLordFrameConfig = gLordFrameConfig

	gLordChatConfig := make(map[int]*t.LordChatConfig)
	loadConfig(path, "LordChatConfig", &gLordChatConfig)
	GLordChatConfig = gLordChatConfig

	gChangeJobConfig := make(map[int]*t.ChangeJobConfig)
	loadConfig(path, "ChangeJobConfig", &gChangeJobConfig)
	GChangeJobConfig = gChangeJobConfig

	gItemConfig := make(map[int]*t.ItemConfig)
	loadConfig(path, "ItemConfig", &gItemConfig)
	GItemConfig = gItemConfig

	gItemGroupConfig := make(map[int]map[int]*t.ItemGroupConfig)
	loadConfig(path, "ItemGroupConfig", &gItemGroupConfig)
	GItemGroupConfig = gItemGroupConfig

	gVirtualCurrencyConfig := make(map[int]*t.VirtualCurrencyConfig)
	loadConfig(path, "VirtualCurrencyConfig", &gVirtualCurrencyConfig)
	GVirtualCurrencyConfig = gVirtualCurrencyConfig

	gTalentConfig := make(map[int]*t.TalentConfig)
	loadConfig(path, "TalentConfig", &gTalentConfig)
	GTalentConfig = gTalentConfig

	gTalentLevelConfig := make(map[int]map[int]*t.TalentLevelConfig)
	loadConfig(path, "TalentLevelConfig", &gTalentLevelConfig)
	GTalentLevelConfig = gTalentLevelConfig

	gLordSkillConfig := make(map[int]map[int]*t.LordSkillConfig)
	loadConfig(path, "LordSkillConfig", &gLordSkillConfig)
	GLordSkillConfig = gLordSkillConfig

	gLordSkillLevelConfig := make(map[int]map[int]map[int]*t.LordSkillLevelConfig)
	loadConfig(path, "LordSkillLevelConfig", &gLordSkillLevelConfig)
	GLordSkillLevelConfig = gLordSkillLevelConfig

	gLordSkillCostConfig := make(map[int]*t.LordSkillCostConfig)
	loadConfig(path, "LordSkillCostConfig", &gLordSkillCostConfig)
	GLordSkillCostConfig = gLordSkillCostConfig

	gGuardLevelConfig := make(map[int]*t.GuardLevelConfig)
	loadConfig(path, "GuardLevelConfig", &gGuardLevelConfig)
	GGuardLevelConfig = gGuardLevelConfig

	gGuardModelConfig := make(map[int]*t.GuardModelConfig)
	loadConfig(path, "GuardModelConfig", &gGuardModelConfig)
	GGuardModelConfig = gGuardModelConfig

	gHeroBaseConfig := &t.HeroBaseConfig{}
	loadConfig(path, "HeroBaseConfig", gHeroBaseConfig)
	GHeroBaseConfig = gHeroBaseConfig

	gHeroConfig := make(map[int]*t.HeroConfig)
	loadConfig(path, "HeroConfig", &gHeroConfig)
	GHeroConfig = gHeroConfig

	gHeroRawConfig := make(map[int]*t.HeroRawConfig)
	loadConfig(path, "HeroRawConfig", &gHeroRawConfig)
	GHeroRawConfig = gHeroRawConfig

	gHeroRarityConfig := make(map[int]*t.HeroRarityConfig)
	loadConfig(path, "HeroRarityConfig", &gHeroRarityConfig)
	GHeroRarityConfig = gHeroRarityConfig

	gHeroLevelConfig := make(map[int]map[int]*t.HeroLevelConfig)
	loadConfig(path, "HeroLevelConfig", &gHeroLevelConfig)
	GHeroLevelConfig = gHeroLevelConfig

	gHeroStageConfig := make(map[int]*t.HeroStageConfig)
	loadConfig(path, "HeroStageConfig", &gHeroStageConfig)
	GHeroStageConfig = gHeroStageConfig

	gHeroStageTalentConfig := make(map[int]map[int]*t.HeroStageTalentConfig)
	loadConfig(path, "HeroStageTalentConfig", &gHeroStageTalentConfig)
	GHeroStageTalentConfig = gHeroStageTalentConfig

	gHeroChangeJobConfig := make(map[int]map[int]*t.HeroChangeJobConfig)
	loadConfig(path, "HeroChangeJobConfig", &gHeroChangeJobConfig)
	GHeroChangeJobConfig = gHeroChangeJobConfig

	gHeroChangeJobCostConfig := make(map[int]map[int]*t.HeroChangeJobCostConfig)
	loadConfig(path, "HeroChangeJobCostConfig", &gHeroChangeJobCostConfig)
	GHeroChangeJobCostConfig = gHeroChangeJobCostConfig

	gHeroFetterConfig := make(map[int]map[int]*t.HeroFetterConfig)
	loadConfig(path, "HeroFetterConfig", &gHeroFetterConfig)
	GHeroFetterConfig = gHeroFetterConfig

	gHeroOpenConfig := make(map[int]map[int]*t.HeroOpenConfig)
	loadConfig(path, "HeroOpenConfig", &gHeroOpenConfig)
	GHeroOpenConfig = gHeroOpenConfig

	gMonsterConfig := make(map[int]*t.MonsterConfig)
	loadConfig(path, "MonsterConfig", &gMonsterConfig)
	GMonsterConfig = gMonsterConfig

	gMonsterTemplateConfig := make(map[int]map[int]*t.MonsterTemplateConfig)
	loadConfig(path, "MonsterTemplateConfig", &gMonsterTemplateConfig)
	GMonsterTemplateConfig = gMonsterTemplateConfig

	gMonsterLordConfig := make(map[int]*t.MonsterLordConfig)
	loadConfig(path, "MonsterLordConfig", &gMonsterLordConfig)
	GMonsterLordConfig = gMonsterLordConfig

	gMonsterLordTemplateConfig := make(map[int]map[int]*t.MonsterLordTemplateConfig)
	loadConfig(path, "MonsterLordTemplateConfig", &gMonsterLordTemplateConfig)
	GMonsterLordTemplateConfig = gMonsterLordTemplateConfig

	gMainFubenBaseConfig := &t.MainFubenBaseConfig{}
	loadConfig(path, "MainFubenBaseConfig", gMainFubenBaseConfig)
	GMainFubenBaseConfig = gMainFubenBaseConfig

	gMainFubenConfig := make(map[int]*t.MainFubenConfig)
	loadConfig(path, "MainFubenConfig", &gMainFubenConfig)
	GMainFubenConfig = gMainFubenConfig

	gBuffConfig := make(map[int]*t.BuffConfig)
	loadConfig(path, "BuffConfig", &gBuffConfig)
	GBuffConfig = gBuffConfig

	gBuffBaseConfig := make(map[int]*t.BuffBaseConfig)
	loadConfig(path, "BuffBaseConfig", &gBuffBaseConfig)
	GBuffBaseConfig = gBuffBaseConfig

	gSkillConfig := make(map[int]*t.SkillConfig)
	loadConfig(path, "SkillConfig", &gSkillConfig)
	GSkillConfig = gSkillConfig

	gSkillEffectConfig := make(map[int]map[int]*t.SkillEffectConfig)
	loadConfig(path, "SkillEffectConfig", &gSkillEffectConfig)
	GSkillEffectConfig = gSkillEffectConfig

	gRaceConfig := make(map[int]*t.RaceConfig)
	loadConfig(path, "RaceConfig", &gRaceConfig)
	GRaceConfig = gRaceConfig

	gEquipBaseConfig := &t.EquipBaseConfig{}
	loadConfig(path, "EquipBaseConfig", gEquipBaseConfig)
	GEquipBaseConfig = gEquipBaseConfig

	gEquipConfig := make(map[int]*t.EquipConfig)
	loadConfig(path, "EquipConfig", &gEquipConfig)
	GEquipConfig = gEquipConfig

	gEquipPosConfig := make(map[int]*t.EquipPosConfig)
	loadConfig(path, "EquipPosConfig", &gEquipPosConfig)
	GEquipPosConfig = gEquipPosConfig

	gEquipRarityConfig := make(map[int]*t.EquipRarityConfig)
	loadConfig(path, "EquipRarityConfig", &gEquipRarityConfig)
	GEquipRarityConfig = gEquipRarityConfig

	gEquipStrengConfig := make(map[int]map[int]*t.EquipStrengConfig)
	loadConfig(path, "EquipStrengConfig", &gEquipStrengConfig)
	GEquipStrengConfig = gEquipStrengConfig

	gEquipSuitConfig := make(map[int]*t.EquipSuitConfig)
	loadConfig(path, "EquipSuitConfig", &gEquipSuitConfig)
	GEquipSuitConfig = gEquipSuitConfig

	gEquipStrengMasterConfig := make(map[int]*t.EquipStrengMasterConfig)
	loadConfig(path, "EquipStrengMasterConfig", &gEquipStrengMasterConfig)
	GEquipStrengMasterConfig = gEquipStrengMasterConfig

	gArtifactBaseConfig := &t.ArtifactBaseConfig{}
	loadConfig(path, "ArtifactBaseConfig", gArtifactBaseConfig)
	GArtifactBaseConfig = gArtifactBaseConfig

	gArtifactConfig := make(map[int]*t.ArtifactConfig)
	loadConfig(path, "ArtifactConfig", &gArtifactConfig)
	GArtifactConfig = gArtifactConfig

	gArtifactRarityConfig := make(map[int]map[int]*t.ArtifactRarityConfig)
	loadConfig(path, "ArtifactRarityConfig", &gArtifactRarityConfig)
	GArtifactRarityConfig = gArtifactRarityConfig

	gArtifactStrengConfig := make(map[int]map[int]*t.ArtifactStrengConfig)
	loadConfig(path, "ArtifactStrengConfig", &gArtifactStrengConfig)
	GArtifactStrengConfig = gArtifactStrengConfig

	gBuildingConfig := make(map[int]*t.BuildingConfig)
	loadConfig(path, "BuildingConfig", &gBuildingConfig)
	GBuildingConfig = gBuildingConfig

	gBuildingLevelConfig := make(map[int]map[int]*t.BuildingLevelConfig)
	loadConfig(path, "BuildingLevelConfig", &gBuildingLevelConfig)
	GBuildingLevelConfig = gBuildingLevelConfig

	gChatBaseConfig := &t.ChatBaseConfig{}
	loadConfig(path, "ChatBaseConfig", gChatBaseConfig)
	GChatBaseConfig = gChatBaseConfig

	gSystemNoticeConfig := make(map[int]map[int]map[int]*t.SystemNoticeConfig)
	loadConfig(path, "SystemNoticeConfig", &gSystemNoticeConfig)
	GSystemNoticeConfig = gSystemNoticeConfig

	gSystemTimingNoticeConfig := make(map[int]*t.SystemTimingNoticeConfig)
	loadConfig(path, "SystemTimingNoticeConfig", &gSystemTimingNoticeConfig)
	GSystemTimingNoticeConfig = gSystemTimingNoticeConfig

	gSystemConfig := &t.SystemConfig{}
	loadConfig(path, "SystemConfig", gSystemConfig)
	GSystemConfig = gSystemConfig

	gSystemOpenConfig := make(map[int]*t.SystemOpenConfig)
	loadConfig(path, "SystemOpenConfig", &gSystemOpenConfig)
	GSystemOpenConfig = gSystemOpenConfig

	gFightSkipConfig := make(map[int]*t.FightSkipConfig)
	loadConfig(path, "FightSkipConfig", &gFightSkipConfig)
	GFightSkipConfig = gFightSkipConfig

	gWorldBossBaseConfig := &t.WorldBossBaseConfig{}
	loadConfig(path, "WorldBossBaseConfig", gWorldBossBaseConfig)
	GWorldBossBaseConfig = gWorldBossBaseConfig

	gWorldBossConfig := make(map[int]*t.WorldBossConfig)
	loadConfig(path, "WorldBossConfig", &gWorldBossConfig)
	GWorldBossConfig = gWorldBossConfig

	gWorldBossRankConfig := make(map[int]map[int]*t.WorldBossRankConfig)
	loadConfig(path, "WorldBossRankConfig", &gWorldBossRankConfig)
	GWorldBossRankConfig = gWorldBossRankConfig
}
