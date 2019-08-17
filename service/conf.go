package service

import (
	"fmt"
	"github.com/zoenion/common/conf"
	"path/filepath"
)

func (v *Vars) ConfigFilename() string {
	return filepath.Join(v.Dir, "configs.json")
}

func (v *ConfigVars) ConfigFilename() string {
	return filepath.Join(v.Dir, "configs.json")
}

func LoadConfigs(dir string) (conf.Map, error) {
	cfgFilename := filepath.Join(dir, "configs.json")
	cfg := conf.Map{}
	err := cfg.Load(cfgFilename)
	if err != nil {
		return nil, fmt.Errorf("could not load configs: %s", err)
	}
	return cfg, nil
}
