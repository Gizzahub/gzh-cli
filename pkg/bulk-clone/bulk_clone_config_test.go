package bulkclone

import (
	"testing"

	"gopkg.in/yaml.v2"
)

// func (receiver ) name()  {
//
//}

func TestReadConfig(t *testing.T) {
	// use bulk-clone.yaml
	// call setclond_config.ReadConfig
	config := &bulkCloneConfig{}
	// bulkCloneConfig.ReadConfig("../../../test")
	// config.ReadConfig("../../../test")
	if err := config.ReadConfig("./"); err != nil {
		t.Logf("Warning: failed to read config: %v", err)
	}
	// t.Log(yaml.Marshal(config))
	// print unmarshal yaml format
	yamlData, err := yaml.Marshal(&config)
	if err != nil {
		t.Error(err)
	}

	t.Log(string(yamlData))
}
