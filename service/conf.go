package service

import (
	"fmt"
	"github.com/zoenion/common/conf"
	"github.com/zoenion/common/errors"
	configpb "github.com/zoenion/common/proto/config"
	"github.com/zoenion/common/service/net"
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

func configClient(v *Vars) (configpb.ConfigClient, error) {
	if v.configClient != nil {
		return v.configClient, nil
	}

	if v.ConfigServer == "" {
		return nil, errors.NotFound
	}

	conn, err := net.GRPCMutualTlsDial(v.ConfigServer, v.authorityCert, v.serviceCert, v.serviceKey)
	if err != nil {
		return nil, err
	}

	v.configClient = configpb.NewConfigClient(conn)
	return v.configClient, nil
}
