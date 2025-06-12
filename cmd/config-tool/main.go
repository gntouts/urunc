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
	"fmt"
	"os"
	"path/filepath"

	"github.com/davecgh/go-spew/spew"
)

func main() {
	fmt.Println("Hi")
	fname := myConfigFile()
	fmt.Println("Config file path:", fname)
	cfg, err := loadConfig(fname)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}
	spew.Dump(cfg)
}

func myConfigFile() string {
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, "config.toml")
}
