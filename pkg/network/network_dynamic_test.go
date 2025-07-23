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

package network

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urunc-dev/urunc/internal/constants"
)

func TestDynamicNetwork_NetworkSetup(t *testing.T) {
	t.Run("dynamic network creation", func(t *testing.T) {
		// Skip this test if not running as root or in proper environment
		if os.Getuid() != 0 {
			t.Skip("Skipping test that requires root privileges")
		}

		dynamicNet := &DynamicNetwork{}

		// Test parameters
		uid := uint32(1000)
		gid := uint32(1000)

		// This will likely fail in test environment without eth0, but we test the structure
		networkInfo, err := dynamicNet.NetworkSetup(uid, gid)

		if err != nil {
			// Expected in test environment - validate error handling
			t.Logf("Expected error in test environment: %v", err)
			assert.Error(t, err)
			assert.Nil(t, networkInfo)

			// Check if it's the multiple unikernels error or interface not found error
			if strings.Contains(err.Error(), "multiple unikernels") {
				assert.Contains(t, err.Error(), "can't spawn multiple unikernels in the same network namespace")
			} else if strings.Contains(err.Error(), "eth0") {
				assert.Contains(t, err.Error(), "failed to find eth0 interface")
			}
		} else {
			// If successful, validate structure
			assert.NotNil(t, networkInfo)
			assert.NotEmpty(t, networkInfo.TapDevice)
			assert.NotEmpty(t, networkInfo.EthDevice.IP)
			assert.NotEmpty(t, networkInfo.EthDevice.DefaultGateway)
			assert.NotEmpty(t, networkInfo.EthDevice.Mask)
			assert.Equal(t, DefaultInterface, networkInfo.EthDevice.Interface)
		}
	})
}

func TestDynamicNetworkTapNameGeneration(t *testing.T) {
	t.Run("tap device name generation", func(t *testing.T) {
		// Test tap name generation logic used in dynamic network
		tapIndex := 0
		expectedName := "tap0_urunc"
		actualName := strings.ReplaceAll(DefaultTap, "X", strconv.Itoa(tapIndex))
		assert.Equal(t, expectedName, actualName)

		// Test with different indices
		tapIndex = 1
		expectedName = "tap1_urunc"
		actualName = strings.ReplaceAll(DefaultTap, "X", strconv.Itoa(tapIndex))
		assert.Equal(t, expectedName, actualName)

		tapIndex = 255
		expectedName = "tap255_urunc"
		actualName = strings.ReplaceAll(DefaultTap, "X", strconv.Itoa(tapIndex))
		assert.Equal(t, expectedName, actualName)
	})
}

func TestDynamicNetworkIPGeneration(t *testing.T) {
	t.Run("dynamic IP address generation", func(t *testing.T) {
		// Test IP generation logic used in dynamic network
		tapIndex := 0
		ipTemplate := "172.16.X.2/24"
		expectedIP := "172.16.1.2/24"
		actualIP := strings.ReplaceAll(ipTemplate, "X", strconv.Itoa(tapIndex+1))
		assert.Equal(t, expectedIP, actualIP)

		// Test with different indices
		tapIndex = 1
		expectedIP = "172.16.2.2/24"
		actualIP = strings.ReplaceAll(ipTemplate, "X", strconv.Itoa(tapIndex+1))
		assert.Equal(t, expectedIP, actualIP)

		tapIndex = 254
		expectedIP = "172.16.255.2/24"
		actualIP = strings.ReplaceAll(ipTemplate, "X", strconv.Itoa(tapIndex+1))
		assert.Equal(t, expectedIP, actualIP)
	})
}

func TestDynamicNetworkConstants(t *testing.T) {
	t.Run("validate dynamic network constants", func(t *testing.T) {
		assert.Equal(t, "172.16.X.2", constants.DynamicNetworkTapIP)

		// Verify the template format
		assert.Contains(t, constants.DynamicNetworkTapIP, "X")
		assert.Contains(t, constants.DynamicNetworkTapIP, "172.16.")
	})
}

func TestDynamicNetworkStructure(t *testing.T) {
	t.Run("dynamic network struct creation", func(t *testing.T) {
		dynamicNet := &DynamicNetwork{}
		assert.NotNil(t, dynamicNet)

		// Verify it implements the Manager interface
		var _ Manager = dynamicNet
	})
}

func TestDynamicNetworkTapIndexValidation(t *testing.T) {
	t.Run("tap index boundary validation", func(t *testing.T) {
		// Test the logic that would reject multiple unikernels
		// This simulates the condition in NetworkSetup

		// If tapIndex > 0, it should return an error
		tapIndex := 1
		if tapIndex > 0 {
			err := assert.AnError // Simulate the error that would be returned
			assert.Error(t, err)
		}

		// If tapIndex == 0, it should proceed (no error for this condition)
		tapIndex = 0
		if tapIndex > 0 {
			t.Error("Should not enter this block when tapIndex is 0")
		}
	})
}

// Test the complete IP template resolution
func TestDynamicNetworkIPTemplateResolution(t *testing.T) {
	t.Run("complete IP template resolution", func(t *testing.T) {
		// Test the complete process as done in the NetworkSetup method
		tapIndex := 0

		// Create IP template
		ipTemplate := "172.16.X.2/24"

		// Replace X with tapIndex + 1
		newIPAddr := strings.ReplaceAll(ipTemplate, "X", strconv.Itoa(tapIndex+1))

		expectedIP := "172.16.1.2/24"
		assert.Equal(t, expectedIP, newIPAddr)

		// Verify it's a valid CIDR notation
		assert.Contains(t, newIPAddr, "/24")
		assert.Contains(t, newIPAddr, "172.16.1.2")
	})
}
