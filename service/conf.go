package service

import (
	"fmt"
	"github.com/zoenion/common/conf"
	"os"
	"path/filepath"
)

func (v *Vars) ConfigFilename() string {
	return filepath.Join(v.dir, "configs.json")
}

func LoadConfigs(v *Vars) (conf.Map, error) {
	cfgFilename := filepath.Join(v.dir, "configs.json")
	cfg := conf.Map{}
	err := cfg.Load(cfgFilename)
	if err != nil {
		return nil, fmt.Errorf("could not load configs: %s", err)
	}
	return cfg, nil
}

func SaveConfigs(v *Vars, cfg conf.Map) error {
	cfgFilename := filepath.Join(v.dir, "configs.json")
	return cfg.Save(cfgFilename, os.ModePerm)
}

func (v *ConfigVars) ConfigFilename() string {
	return filepath.Join(v.dir, "configs.json")
}

func LoadOldConfigs(v *ConfigVars) (conf.Map, error) {
	cfgFilename := filepath.Join(v.dir, "configs.json")
	cfg := conf.Map{}
	err := cfg.Load(cfgFilename)
	if err != nil {
		return nil, fmt.Errorf("could not load configs: %s", err)
	}
	return cfg, nil
}
