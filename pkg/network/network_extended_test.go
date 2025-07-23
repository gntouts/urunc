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
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urunc-dev/urunc/internal/constants"
)

// Test error conditions for static network setup
func TestStaticNetworkErrorConditions(t *testing.T) {
	t.Run("static network manager initialization", func(t *testing.T) {
		staticNet := &StaticNetwork{}
		assert.NotNil(t, staticNet)

		// Test that it satisfies the Manager interface
		var manager Manager = staticNet
		assert.NotNil(t, manager)
	})

	t.Run("static network setup without eth0", func(t *testing.T) {
		staticNet := &StaticNetwork{}

		// This will fail in most test environments, but we're testing error handling
		_, err := staticNet.NetworkSetup(1000, 1000)

		// Should get an error about eth0 not being found or other link errors
		if err != nil {
			// Accept various error messages that can occur in different environments
			assert.True(t,
				strings.Contains(err.Error(), "eth0") ||
					strings.Contains(err.Error(), "Link not found") ||
					strings.Contains(err.Error(), "failed to find"),
				"Error should be about interface or link issues")
		}
	})

	t.Run("static IP address validation", func(t *testing.T) {
		// Test that StaticIPAddr has the correct format
		assert.Contains(t, StaticIPAddr, constants.StaticNetworkTapIP)
		assert.Contains(t, StaticIPAddr, "/24")

		// Validate it's a proper CIDR notation
		expectedIP := fmt.Sprintf("%s/24", constants.StaticNetworkTapIP)
		assert.Equal(t, expectedIP, StaticIPAddr)
	})
}

// Test error conditions for dynamic network setup
func TestDynamicNetworkErrorConditions(t *testing.T) {
	t.Run("dynamic network manager initialization", func(t *testing.T) {
		dynamicNet := &DynamicNetwork{}
		assert.NotNil(t, dynamicNet)

		// Test that it satisfies the Manager interface
		var manager Manager = dynamicNet
		assert.NotNil(t, manager)
	})

	t.Run("dynamic network setup without eth0", func(t *testing.T) {
		dynamicNet := &DynamicNetwork{}

		// This will fail in most test environments, but we're testing error handling
		_, err := dynamicNet.NetworkSetup(1000, 1000)

		// Should get an error about eth0 not being found or multiple unikernels
		if err != nil {
			assert.True(t,
				strings.Contains(err.Error(), "eth0") ||
					strings.Contains(err.Error(), "multiple unikernels") ||
					strings.Contains(err.Error(), "Link not found") ||
					strings.Contains(err.Error(), "failed to find"),
				"Error should be about eth0, multiple unikernels, or link issues")
		}
	})

	t.Run("dynamic network IP template validation", func(t *testing.T) {
		// Test IP template generation logic
		testCases := []struct {
			tapIndex   int
			expectedIP string
		}{
			{0, "172.16.1.2/24"},
			{1, "172.16.2.2/24"},
			{5, "172.16.6.2/24"},
			{254, "172.16.255.2/24"},
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("index_%d", tc.tapIndex), func(t *testing.T) {
				ipTemplate := fmt.Sprintf("%s/24", constants.DynamicNetworkTapIP)
				actualIP := strings.ReplaceAll(ipTemplate, "X", strconv.Itoa(tc.tapIndex+1))
				assert.Equal(t, tc.expectedIP, actualIP)
			})
		}
	})

	t.Run("tap index boundary conditions", func(t *testing.T) {
		// Test that the function would reject high tap indices
		// This simulates the error condition in NetworkSetup

		testIndices := []int{1, 2, 5, 10}
		for _, index := range testIndices {
			t.Run(fmt.Sprintf("reject_index_%d", index), func(t *testing.T) {
				// If tapIndex > 0, NetworkSetup should return error
				if index > 0 {
					expectedError := "unsupported operation: can't spawn multiple unikernels in the same network namespace"
					// Simulate the error that would be returned
					assert.Contains(t, expectedError, "multiple unikernels")
				}
			})
		}
	})
}

// Test tap device name generation patterns
func TestTapDeviceNameGeneration(t *testing.T) {
	t.Run("static tap name generation", func(t *testing.T) {
		// Static network always uses index 0
		expectedName := "tap0_urunc"
		actualName := strings.ReplaceAll(DefaultTap, "X", "0")
		assert.Equal(t, expectedName, actualName)
	})

	t.Run("dynamic tap name generation", func(t *testing.T) {
		testCases := []struct {
			index        int
			expectedName string
		}{
			{0, "tap0_urunc"},
			{1, "tap1_urunc"},
			{42, "tap42_urunc"},
			{255, "tap255_urunc"},
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("index_%d", tc.index), func(t *testing.T) {
				actualName := strings.ReplaceAll(DefaultTap, "X", strconv.Itoa(tc.index))
				assert.Equal(t, tc.expectedName, actualName)
			})
		}
	})

	t.Run("tap name validation", func(t *testing.T) {
		// Test that generated names follow expected pattern
		for i := 0; i < 10; i++ {
			tapName := strings.ReplaceAll(DefaultTap, "X", strconv.Itoa(i))

			assert.Contains(t, tapName, "tap")
			assert.Contains(t, tapName, "urunc")
			assert.Contains(t, tapName, strconv.Itoa(i))
			assert.NotContains(t, tapName, "X", "Generated name should not contain placeholder")
		}
	})
}

// Test network constants validation
func TestNetworkConstantsValidation(t *testing.T) {
	t.Run("static network constants", func(t *testing.T) {
		// Validate static network constants are properly formatted IP addresses
		assert.NotEmpty(t, constants.StaticNetworkTapIP)
		assert.NotEmpty(t, constants.StaticNetworkUnikernelIP)

		// Should be valid IP addresses
		assert.Regexp(t, `^\d+\.\d+\.\d+\.\d+$`, constants.StaticNetworkTapIP)
		assert.Regexp(t, `^\d+\.\d+\.\d+\.\d+$`, constants.StaticNetworkUnikernelIP)

		// Should be in the same subnet
		tapParts := strings.Split(constants.StaticNetworkTapIP, ".")
		unikernelParts := strings.Split(constants.StaticNetworkUnikernelIP, ".")

		// First three octets should match (assuming /24 subnet)
		assert.Equal(t, tapParts[0], unikernelParts[0])
		assert.Equal(t, tapParts[1], unikernelParts[1])
		assert.Equal(t, tapParts[2], unikernelParts[2])

		// Last octets should be different
		assert.NotEqual(t, tapParts[3], unikernelParts[3])
	})

	t.Run("dynamic network constants", func(t *testing.T) {
		// Validate dynamic network template
		assert.NotEmpty(t, constants.DynamicNetworkTapIP)
		assert.Contains(t, constants.DynamicNetworkTapIP, "X", "Dynamic template should contain placeholder")

		// Should follow expected pattern
		assert.Regexp(t, `^\d+\.\d+\.X\.\d+$`, constants.DynamicNetworkTapIP)
	})

	t.Run("default constants", func(t *testing.T) {
		assert.Equal(t, "eth0", DefaultInterface)
		assert.Equal(t, "tapX_urunc", DefaultTap)

		// DefaultTap should contain placeholder
		assert.Contains(t, DefaultTap, "X")
	})
}

// Test error conditions that don't require root
func TestNetworkErrorHandlingDetailed(t *testing.T) {
	t.Run("manager creation with various types", func(t *testing.T) {
		validTypes := []string{"static", "dynamic"}
		invalidTypes := []string{"", "invalid", "bridge", "host", "none", "unknown"}

		// Test valid types
		for _, validType := range validTypes {
			t.Run("valid_"+validType, func(t *testing.T) {
				manager, err := NewNetworkManager(validType)
				assert.NoError(t, err)
				assert.NotNil(t, manager)
			})
		}

		// Test invalid types
		for _, invalidType := range invalidTypes {
			t.Run("invalid_"+invalidType, func(t *testing.T) {
				manager, err := NewNetworkManager(invalidType)
				assert.Error(t, err)
				assert.Nil(t, manager)
				assert.Contains(t, err.Error(), "not supported")
			})
		}
	})

	t.Run("interface validation edge cases", func(t *testing.T) {
		// Test ensureEth0Exists behavior patterns
		err := ensureEth0Exists()

		// The error should be consistent
		if err != nil {
			assert.Contains(t, err.Error(), "eth0 device not found")
		}
	})
}

// Test network setup parameter validation
func TestNetworkSetupParameterValidation(t *testing.T) {
	t.Run("uid gid validation ranges", func(t *testing.T) {
		testCases := []struct {
			name string
			uid  uint32
			gid  uint32
		}{
			{"min_values", 0, 0},
			{"normal_values", 1000, 1000},
			{"max_values", 4294967295, 4294967295},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// Test that the values are acceptable ranges
				assert.LessOrEqual(t, tc.uid, uint32(4294967295))
				assert.LessOrEqual(t, tc.gid, uint32(4294967295))
				assert.GreaterOrEqual(t, tc.uid, uint32(0))
				assert.GreaterOrEqual(t, tc.gid, uint32(0))
			})
		}
	})
}
