// Copyright (c) 2020-2021 Doc.ai and/or its affiliates.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package config provides methods to get configuration parameters from environment variables
package config

import (
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"

	"github.com/networkservicemesh/api/pkg/api/networkservice/payload"
)

const (
	addrPrefix    = "addr:"
	vlanPrefix    = "vlan:"
	labelsPrefix  = "labels:"
	payloadPrefix = "payload:"
)

// Config holds configuration parameters from environment variables
type Config struct {
	Name                      string        `default:"vfio-server" desc:"name of VFIO Server" split_words:"true"`
	BaseDir                   string        `default:"./" desc:"base directory" split_words:"true"`
	ConnectTo                 url.URL       `default:"unix:///var/lib/networkservicemesh/nsm.io.sock" desc:"url to connect to" split_words:"true"`
	MaxTokenLifetime          time.Duration `default:"10m" desc:"maximum lifetime of tokens" split_words:"true"`
	LogLevel                  string        `default:"INFO" desc:"Log level" split_words:"true"`
	OpenTelemetryCollectorURL string        `default:"otel-collector.observability.svc.cluster.local:4317" desc:"OpenTelemetry Collector URL"`

	Services []ServiceConfig `default:"" desc:"list of supported services"`
}

// Process prints and processes env to config
func (c *Config) Process() error {
	if err := envconfig.Usage("nsm", c); err != nil {
		return errors.Wrap(err, "cannot show usage of envconfig nse")
	}
	if err := envconfig.Process("nsm", c); err != nil {
		return errors.Wrap(err, "cannot process envconfig nse")
	}
	return nil
}

// ServiceConfig is a per-service config
type ServiceConfig struct {
	Name    string
	Domain  string
	Payload string
	MACAddr net.HardwareAddr
	VLANTag int32
	Labels  map[string]string
}

// UnmarshalBinary expects string(bytes) to be in format:
// Name@Domain: { addr: MACAddr; vlan: VLANTag; labels: Labels; payload: Payload; }
// MACAddr = xx:xx:xx:xx:xx:xx
// Labels = label_1=value_1&label_2=value_2
func (s *ServiceConfig) UnmarshalBinary(bytes []byte) (err error) {
	text := string(bytes)

	split := strings.Split(text, "@")
	s.Name = strings.TrimSpace(split[0])

	if len(split) < 2 {
		return errors.Errorf("invalid format: %s", text)
	}

	split = strings.Split(split[1], ":")
	s.Domain = strings.TrimSpace(split[0])

	split = strings.Split(text, "{")
	if len(split) < 2 {
		return errors.Errorf("invalid format: %s", text)
	}

	// Set default Payload
	s.Payload = payload.Ethernet

	split = strings.Split(split[1], "}")
	for _, part := range strings.Split(split[0], ";") {
		part = strings.TrimSpace(part)
		switch {
		case strings.HasPrefix(part, addrPrefix):
			s.MACAddr, err = net.ParseMAC(trimPrefix(part, addrPrefix))
		case strings.HasPrefix(part, vlanPrefix):
			s.VLANTag, err = parseInt32(trimPrefix(part, vlanPrefix))
		case strings.HasPrefix(part, labelsPrefix):
			s.Labels, err = parseMap(trimPrefix(part, labelsPrefix))
		case strings.HasPrefix(part, payloadPrefix):
			s.Payload = trimPrefix(part, payloadPrefix)
		default:
			err = errors.Errorf("invalid format: %s", text)
		}
		if err != nil {
			return err
		}
	}

	return s.validate()
}

func trimPrefix(s, prefix string) string {
	s = strings.TrimPrefix(s, prefix)
	return strings.TrimSpace(s)
}

func parseInt32(s string) (int32, error) {
	i, err := strconv.ParseInt(s, 0, 32)
	if err != nil {
		return 0, err
	}
	return int32(i), nil
}

func parseMap(s string) (map[string]string, error) {
	m := make(map[string]string)
	for _, keyValue := range strings.Split(s, "&") {
		split := strings.Split(keyValue, "=")
		if len(split) != 2 {
			return nil, errors.Errorf("invalid key-value pair: %s", keyValue)
		}
		m[split[0]] = split[1]
	}
	return m, nil
}

func (s *ServiceConfig) validate() error {
	if s.Name == "" {
		return errors.New("name is empty")
	}
	if s.Domain == "" {
		return errors.New("domain is empty")
	}
	if s.MACAddr.String() == "" {
		return errors.New("MAC address is empty")
	}
	return nil
}
