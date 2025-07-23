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
	"net"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewNetworkManager(t *testing.T) {
	tests := []struct {
		name         string
		networkType  string
		expectedType string
		expectError  bool
		errorMessage string
	}{
		{
			name:         "static network manager",
			networkType:  "static",
			expectedType: "*network.StaticNetwork",
			expectError:  false,
		},
		{
			name:         "dynamic network manager",
			networkType:  "dynamic",
			expectedType: "*network.DynamicNetwork",
			expectError:  false,
		},
		{
			name:         "unsupported network manager",
			networkType:  "unsupported",
			expectError:  true,
			errorMessage: "network manager unsupported not supported",
		},
		{
			name:         "empty network type",
			networkType:  "",
			expectError:  true,
			errorMessage: "network manager  not supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewNetworkManager(tt.networkType)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, manager)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, manager)
				assert.Equal(t, tt.expectedType, reflect.TypeOf(manager).String())
			}
		})
	}
}

func TestGetTapIndex(t *testing.T) {
	// This test requires network interfaces to be available
	// We'll test the function behavior rather than mock network interfaces
	t.Run("get tap index success", func(t *testing.T) {
		index, err := getTapIndex()

		// Should not return error
		assert.NoError(t, err)

		// Index should be non-negative and within valid range
		assert.GreaterOrEqual(t, index, 0)
		assert.LessOrEqual(t, index, 255)
	})
}

func TestInterface(t *testing.T) {
	t.Run("interface struct creation", func(t *testing.T) {
		iface := Interface{
			IP:             "192.168.1.100",
			DefaultGateway: "192.168.1.1",
			Mask:           "255.255.255.0",
			Interface:      "eth0",
			MAC:            "00:11:22:33:44:55",
		}

		assert.Equal(t, "192.168.1.100", iface.IP)
		assert.Equal(t, "192.168.1.1", iface.DefaultGateway)
		assert.Equal(t, "255.255.255.0", iface.Mask)
		assert.Equal(t, "eth0", iface.Interface)
		assert.Equal(t, "00:11:22:33:44:55", iface.MAC)
	})
}

func TestUnikernelNetworkInfo(t *testing.T) {
	t.Run("unikernel network info creation", func(t *testing.T) {
		ethDevice := Interface{
			IP:             "10.0.0.2",
			DefaultGateway: "10.0.0.1",
			Mask:           "255.255.255.0",
			Interface:      "eth0",
			MAC:            "aa:bb:cc:dd:ee:ff",
		}

		networkInfo := UnikernelNetworkInfo{
			TapDevice: "tap0_urunc",
			EthDevice: ethDevice,
		}

		assert.Equal(t, "tap0_urunc", networkInfo.TapDevice)
		assert.Equal(t, ethDevice, networkInfo.EthDevice)
	})
}

func TestEnsureEth0Exists(t *testing.T) {
	t.Run("check eth0 existence", func(t *testing.T) {
		err := ensureEth0Exists()

		// This will either return nil if eth0 exists or an error if it doesn't
		// We can't guarantee eth0 exists in test environment, so we just verify
		// the function doesn't panic and returns appropriate result
		if err != nil {
			assert.Contains(t, err.Error(), "eth0 device not found")
		}
	})
}

func TestGetInterfaceInfo(t *testing.T) {
	t.Run("get loopback interface info", func(t *testing.T) {
		// Test with loopback interface which should exist on most systems
		info, err := getInterfaceInfo("lo")

		if err != nil {
			// If we get an error, it should be a reasonable one
			t.Logf("Expected error for loopback interface: %v", err)
		} else {
			// If successful, validate the structure
			assert.NotEmpty(t, info.Interface)
			assert.Equal(t, "lo", info.Interface)
		}
	})

	t.Run("get non-existent interface info", func(t *testing.T) {
		_, err := getInterfaceInfo("nonexistent999")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no such network interface")
	})
}

// Test helper function to check if we can list network interfaces
func TestNetworkInterfacesAvailable(t *testing.T) {
	t.Run("list network interfaces", func(t *testing.T) {
		interfaces, err := net.Interfaces()
		require.NoError(t, err)
		assert.Greater(t, len(interfaces), 0, "Expected at least one network interface")

		// Log available interfaces for debugging
		for _, iface := range interfaces {
			t.Logf("Available interface: %s", iface.Name)
		}
	})
}

// Test constants
func TestConstants(t *testing.T) {
	t.Run("default constants", func(t *testing.T) {
		assert.Equal(t, "eth0", DefaultInterface)
		assert.Equal(t, "tapX_urunc", DefaultTap)
	})
}
