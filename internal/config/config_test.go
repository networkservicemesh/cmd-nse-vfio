package config_test

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/networkservicemesh/cmd-nse-vfio/internal/config"
)

func TestServiceConfig_UnmarshalBinary(t *testing.T) {
	cfg := new(config.ServiceConfig)

	err := cfg.UnmarshalBinary([]byte("pingpong@worker.domain: { addr: 0a:55:44:33:22:11 }"))
	require.NoError(t, err)

	require.Equal(t, &config.ServiceConfig{
		Name:    "pingpong",
		Domain:  "worker.domain",
		MACAddr: net.HardwareAddr{0x0a, 0x55, 0x44, 0x33, 0x22, 0x11},
	}, cfg)
}
