// nsdp_test.go
//
// Copyright (C) 2022 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
//

package nsdp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	plugin := NewNSDP()
	require.NotNil(t, plugin)
}

func TestSampleConfig(t *testing.T) {
	plugin := NewNSDP()
	sampleConfig := plugin.SampleConfig()
	require.NotNil(t, sampleConfig)
}

func TestDescription(t *testing.T) {
	plugin := NewNSDP()
	description := plugin.Description()
	require.NotNil(t, description)
}
