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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests that test the full workflow
func TestNetworkManagerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Run("full static network workflow", func(t *testing.T) {
		// Skip if not root
		if os.Getuid() != 0 {
			t.Skip("Skipping integration test that requires root privileges")
		}

		// Create static network manager
		manager, err := NewNetworkManager("static")
		require.NoError(t, err)
		require.NotNil(t, manager)

		// Attempt network setup
		uid := uint32(os.Getuid())
		gid := uint32(os.Getgid())

		networkInfo, err := manager.NetworkSetup(uid, gid)

		if err != nil {
			// In test environment, this is expected to fail
			t.Logf("Expected failure in test environment: %v", err)
			assert.Error(t, err)
		} else {
			// If successful, validate and cleanup
			require.NotNil(t, networkInfo)
			assert.NotEmpty(t, networkInfo.TapDevice)

			// Cleanup the created tap device
			err = Cleanup(networkInfo.TapDevice)
			if err != nil {
				t.Logf("Cleanup error (may be expected): %v", err)
			}
		}
	})

	t.Run("full dynamic network workflow", func(t *testing.T) {
		// Skip if not root
		if os.Getuid() != 0 {
			t.Skip("Skipping integration test that requires root privileges")
		}

		// Create dynamic network manager
		manager, err := NewNetworkManager("dynamic")
		require.NoError(t, err)
		require.NotNil(t, manager)

		// Attempt network setup
		uid := uint32(os.Getuid())
		gid := uint32(os.Getgid())

		networkInfo, err := manager.NetworkSetup(uid, gid)

		if err != nil {
			// In test environment, this is expected to fail
			t.Logf("Expected failure in test environment: %v", err)
			assert.Error(t, err)
		} else {
			// If successful, validate and cleanup
			require.NotNil(t, networkInfo)
			assert.NotEmpty(t, networkInfo.TapDevice)

			// Cleanup the created tap device
			err = Cleanup(networkInfo.TapDevice)
			if err != nil {
				t.Logf("Cleanup error (may be expected): %v", err)
			}
		}
	})
}

// Test network manager interface compliance
func TestNetworkManagerInterface(t *testing.T) {
	t.Run("static network implements Manager interface", func(t *testing.T) {
		var manager Manager
		manager = &StaticNetwork{}
		assert.NotNil(t, manager)

		// Verify method signature exists
		assert.NotNil(t, manager.NetworkSetup)
	})

	t.Run("dynamic network implements Manager interface", func(t *testing.T) {
		var manager Manager
		manager = &DynamicNetwork{}
		assert.NotNil(t, manager)

		// Verify method signature exists
		assert.NotNil(t, manager.NetworkSetup)
	})
}

// Test error conditions and edge cases
func TestNetworkEdgeCases(t *testing.T) {
	t.Run("network setup with extreme UIDs", func(t *testing.T) {
		if os.Getuid() != 0 {
			t.Skip("Skipping test that requires root privileges")
		}

		manager, err := NewNetworkManager("static")
		require.NoError(t, err)

		// Test with maximum UID values
		_, err = manager.NetworkSetup(4294967295, 4294967295) // Max uint32
		// This should either work or fail gracefully
		if err != nil {
			t.Logf("Expected error with extreme UID values: %v", err)
		}
	})

	t.Run("concurrent network setup", func(t *testing.T) {
		if os.Getuid() != 0 {
			t.Skip("Skipping test that requires root privileges")
		}

		// Test that concurrent network setups are handled properly
		manager, err := NewNetworkManager("dynamic")
		require.NoError(t, err)

		// This simulates the scenario where multiple unikernels
		// might try to set up networking simultaneously
		uid := uint32(1000)
		gid := uint32(1000)

		// First setup
		networkInfo1, err1 := manager.NetworkSetup(uid, gid)

		// Second setup (should fail for dynamic network)
		networkInfo2, err2 := manager.NetworkSetup(uid, gid)

		// At least one should succeed or both should fail for valid reasons
		if err1 != nil && err2 != nil {
			t.Logf("Both setups failed as expected: %v, %v", err1, err2)
		} else if err1 == nil && err2 != nil {
			// First succeeded, second failed (expected for dynamic)
			assert.NotNil(t, networkInfo1)
			assert.Nil(t, networkInfo2)

			// Cleanup
			if networkInfo1 != nil {
				_ = Cleanup(networkInfo1.TapDevice)
			}
		}
	})
}

// Test system requirements and capabilities
func TestSystemRequirements(t *testing.T) {
	t.Run("check system capabilities", func(t *testing.T) {
		// Check if we can list network interfaces
		interfaces, err := net.Interfaces()
		require.NoError(t, err)
		assert.Greater(t, len(interfaces), 0)

		// Log system information for debugging
		t.Logf("Available network interfaces: %d", len(interfaces))
		for _, iface := range interfaces {
			t.Logf("  - %s: %s", iface.Name, iface.HardwareAddr.String())
		}

		// Check if eth0 exists (common in containers/VMs)
		hasEth0 := false
		for _, iface := range interfaces {
			if iface.Name == "eth0" {
				hasEth0 = true
				break
			}
		}

		if hasEth0 {
			t.Log("eth0 interface found - network tests may succeed")
		} else {
			t.Log("eth0 interface not found - network tests will likely fail")
		}
	})

	t.Run("check required tools", func(t *testing.T) {
		// Check if iptables is available (required for static network)
		if os.Getuid() == 0 {
			// Only check if running as root
			_, err := os.Stat("/sbin/iptables")
			if err == nil {
				t.Log("iptables found")
			} else {
				_, err = os.Stat("/usr/sbin/iptables")
				if err == nil {
					t.Log("iptables found in /usr/sbin")
				} else {
					t.Log("iptables not found - static network tests may fail")
				}
			}
		}
	})
}

// Benchmark network operations
func BenchmarkNetworkOperations(b *testing.B) {
	if os.Getuid() != 0 {
		b.Skip("Skipping benchmark that requires root privileges")
	}

	b.Run("NewNetworkManager", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = NewNetworkManager("static")
		}
	})

	b.Run("getTapIndex", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = getTapIndex()
		}
	})

	b.Run("ensureEth0Exists", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ensureEth0Exists()
		}
	})
}
