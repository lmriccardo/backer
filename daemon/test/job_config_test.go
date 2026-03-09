package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/lmriccardo/backer/deamon/internal/core/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
This tests if the job configuration is loaded correctly
from either a JSON file or JSON string.
*/

func checkConfig(c *domain.JobConfig, t *testing.T) {
	assert.NotEqual(t, c.Name, "", "Name field must not be empty")
	assert.NotEqual(t, c.Log, "", "Log field must not be empty")
	assert.False(t, c.Compression, "Compression must be set to False")
	assert.Len(t, c.Notify, 2, "There must be two notification systems")
}

func TestLoadConfiguration(t *testing.T) {
	abs_conf_path, _ := filepath.Abs("testdata/backup-plan-example.json")
	iostream, err := os.Open(abs_conf_path)
	require.Nilf(t, err, "open config file %s : %v", abs_conf_path, err)

	var job_config domain.JobConfig
	json_encoder := json.NewDecoder(iostream)
	if err := json_encoder.Decode(&job_config); err != nil {
		t.Fatalf("JSON Loading Error: %v", err)
	}

	checkConfig(&job_config, t)
}
