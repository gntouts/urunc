// Copyright (c) 2023-2025, Nubificus LTD
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	GenericRuntime GenericRuntime `toml:"generic_runtime"`
	Hypervisors    Hypervisors    `toml:"hypervisors"`
	RootFS         RootFS         `toml:"rootfs"`
	Logging        Logging        `toml:"logging"`
}

type GenericRuntime struct {
	Path string `toml:"path"`
}

type CgroupsConfig struct {
	UseCgroups bool     `toml:"use_cgroups"`
	Root       string   `toml:"root"`
	Subsystems []string `toml:"subsystems"`
}

type Hypervisors struct {
	Default       string      `toml:"default"`
	DefaultMemory string      `toml:"default_memory"`
	QEMU          *Hypervisor `toml:"qemu,omitempty"`
	Firecracker   *Hypervisor `toml:"firecracker,omitempty"`
	SPT           *Hypervisor `toml:"solo5-spt,omitempty"`
	HVT           *Hypervisor `toml:"solo5-hvt,omitempty"`
}

type Hypervisor struct {
	Path    string            `toml:"path"`
	Options HypervisorOptions `toml:"options"`
}

type HypervisorOptions struct {
	DefaultMemory string   `toml:"default_memory"`
	CliOptions    []string `toml:"cli_options"`
}

type RootFS struct {
	UseDevmapper bool `toml:"use_devmapper"`
	UseSharedFS  bool `toml:"use_sharedfs"`
}

type Logging struct {
	Level           string `toml:"level"`
	Syslog          bool   `toml:"syslog"`
	Timestamp       bool   `toml:"timestamp"`
	TimestampOutput string `toml:"timestamp_output"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	applyDefaults(&cfg)
	return &cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Hypervisors.DefaultMemory == "" {
		cfg.Hypervisors.DefaultMemory = "1G"
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.TimestampOutput == "" {
		cfg.Logging.TimestampOutput = "/var/log/default.log"
	}
	if cfg.GenericRuntime.Path == "" {
		cfg.GenericRuntime.Path = "/usr/bin/default-runtime"
	}
}
