// Teleport
// Copyright (C) 2025 Gravitational, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package vnet

import (
	"context"

	"github.com/gravitational/trace"

	vnetv1 "github.com/gravitational/teleport/gen/proto/go/teleport/lib/vnet/v1"
)

// UserProcessConfig provides the necessary configuration to run VNet.
type UserProcessConfig struct {
	// ClientApplication is a required field providing an interface implementation for
	// [ClientApplication].
	ClientApplication ClientApplication
}

func (c *UserProcessConfig) checkAndSetDefaults() error {
	if c.ClientApplication == nil {
		return trace.BadParameter("missing ClientApplication")
	}
	return nil
}

// RunUserProcess is called by all VNet client applications (tsh, Connect) to
// start and run all VNet tasks.  It returns a [ProcessManager] which controls
// the lifecycle of all tasks and background processes.
//
// ctx is used for setup steps that happen before RunUserProcess passes control
// to the process manager. Canceling ctx after RunUserProcess returns will _not_
// cancel the background tasks. If [RunUserProcess] returns without error, the
// caller is expected to call Close on the process manager to clean up any
// resources, terminate all processes, and remove any OS configuration used for
// actively running VNet.
func RunUserProcess(ctx context.Context, cfg *UserProcessConfig) (*ProcessManager, *vnetv1.NetworkStackInfo, error) {
	if err := cfg.checkAndSetDefaults(); err != nil {
		return nil, nil, trace.Wrap(err)
	}
	return runPlatformUserProcess(ctx, cfg)
}
