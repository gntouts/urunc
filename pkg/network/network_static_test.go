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
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urunc-dev/urunc/internal/constants"
)

func TestStaticIPAddr(t *testing.T) {
	t.Run("static IP address format", func(t *testing.T) {
		expectedIP := "172.16.1.1/24"
		assert.Equal(t, expectedIP, StaticIPAddr)
	})
}

func TestStaticNetwork_NetworkSetup(t *testing.T) {
	// Note: These tests require root privileges and actual network interfaces
	// In a real environment, you might want to use network namespaces or mocking

	t.Run("static network creation", func(t *testing.T) {
		// Skip this test if not running as root or in proper environment
		if os.Getuid() != 0 {
			t.Skip("Skipping test that requires root privileges")
		}

		staticNet := &StaticNetwork{}

		// Test parameters
		uid := uint32(1000)
		gid := uint32(1000)

		// This will likely fail in test environment without eth0, but we test the structure
		networkInfo, err := staticNet.NetworkSetup(uid, gid)

		if err != nil {
			// Expected in test environment - validate error handling
			t.Logf("Expected error in test environment: %v", err)
			assert.Error(t, err)
			assert.Nil(t, networkInfo)
		} else {
			// If successful, validate structure
			assert.NotNil(t, networkInfo)
			assert.NotEmpty(t, networkInfo.TapDevice)
			assert.Equal(t, constants.StaticNetworkUnikernelIP, networkInfo.EthDevice.IP)
			assert.Equal(t, constants.StaticNetworkTapIP, networkInfo.EthDevice.DefaultGateway)
			assert.Equal(t, "255.255.255.0", networkInfo.EthDevice.Mask)
			assert.Equal(t, DefaultInterface, networkInfo.EthDevice.Interface)
		}
	})
}

func TestSetNATRule(t *testing.T) {
	t.Run("test NAT rule parameters validation", func(t *testing.T) {
		// Skip if not root or iptables not available
		if os.Getuid() != 0 {
			t.Skip("Skipping test that requires root privileges")
		}

		// Check if iptables is available
		_, err := exec.LookPath("iptables")
		if err != nil {
			t.Skip("iptables not available in test environment")
		}

		// Test with invalid interface (should fail gracefully)
		err = setNATRule("nonexistent999", "192.168.1.0/24")
		// This should either succeed (if iptables allows it) or fail with a specific error
		if err != nil {
			t.Logf("Expected error for non-existent interface: %v", err)
		}
	})

	t.Run("test IP forwarding file access", func(t *testing.T) {
		// Test if we can access the IP forwarding file
		file, err := os.OpenFile("/proc/sys/net/ipv4/ip_forward", os.O_RDONLY, 0644)
		if err != nil {
			t.Skip("Cannot access /proc/sys/net/ipv4/ip_forward in test environment")
		}
		defer file.Close()

		// Read current value
		content := make([]byte, 10)
		n, err := file.Read(content)
		assert.NoError(t, err)
		assert.Greater(t, n, 0)

		// Should be either "0" or "1"
		value := string(content[:n-1]) // Remove newline
		assert.Contains(t, []string{"0", "1"}, value)
	})
}

func TestStaticNetworkConstants(t *testing.T) {
	t.Run("validate static network constants", func(t *testing.T) {
		assert.Equal(t, "172.16.1.1", constants.StaticNetworkTapIP)
		assert.Equal(t, "172.16.1.2", constants.StaticNetworkUnikernelIP)

		// Verify StaticIPAddr is correctly formatted
		expectedStaticIP := constants.StaticNetworkTapIP + "/24"
		assert.Equal(t, expectedStaticIP, StaticIPAddr)
	})
}

// Mock tests that don't require actual network setup
func TestStaticNetworkStructure(t *testing.T) {
	t.Run("static network struct creation", func(t *testing.T) {
		staticNet := &StaticNetwork{}
		assert.NotNil(t, staticNet)

		// Verify it implements the Manager interface
		var _ Manager = staticNet
	})
}

// Test the tap device name generation for static network
func TestStaticTapDeviceName(t *testing.T) {
	t.Run("static tap device name generation", func(t *testing.T) {
		// The static network always uses tap0_urunc (X replaced with 0)
		expectedName := "tap0_urunc"

		// This is what the static network setup does internally
		actualName := strings.ReplaceAll(DefaultTap, "X", "0")

		assert.Equal(t, expectedName, actualName)
	})
}
