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

// Package mapserver provides chain element implementing `network service -> { MAC, VLAN }` mapping
package mapserver

import (
	"context"
	"net"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"

	"github.com/networkservicemesh/api/pkg/api/networkservice"
	"github.com/networkservicemesh/sdk/pkg/networkservice/core/next"

	"github.com/networkservicemesh/cmd-nse-vfio/internal/config"
)

type mapServer struct {
	entries map[string]*entry
}

type entry struct {
	macAddr net.HardwareAddr
	vlanTag int32
}

// NewServer returns a new `network service -> { MAC, VLAN }` mapping server chain element
func NewServer(cfg *config.Config) networkservice.NetworkServiceServer {
	s := &mapServer{
		entries: make(map[string]*entry, len(cfg.Services)),
	}

	for i := range cfg.Services {
		service := &cfg.Services[i]
		s.entries[service.Name] = &entry{
			macAddr: service.MACAddr,
			vlanTag: service.VLANTag,
		}
	}

	return s
}

func (s *mapServer) Request(ctx context.Context, request *networkservice.NetworkServiceRequest) (*networkservice.Connection, error) {
	conn := request.GetConnection()

	entry, ok := s.entries[conn.GetNetworkService()]
	if !ok {
		return nil, errors.Errorf("network service is not supported: %s", conn.GetNetworkService())
	}

	if conn.GetContext() == nil {
		conn.Context = new(networkservice.ConnectionContext)
	}
	if conn.GetContext().GetEthernetContext() == nil {
		conn.GetContext().EthernetContext = new(networkservice.EthernetContext)
	}
	ethernetContext := conn.GetContext().GetEthernetContext()

	ethernetContext.DstMac = entry.macAddr.String()
	ethernetContext.VlanTag = entry.vlanTag

	return next.Server(ctx).Request(ctx, request)
}

func (s *mapServer) Close(ctx context.Context, conn *networkservice.Connection) (*empty.Empty, error) {
	return next.Server(ctx).Close(ctx, conn)
}
