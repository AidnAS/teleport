/*
 * Teleport
 * Copyright (C) 2023  Gravitational, Inc.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package kafka

import (
	"context"
	"encoding/json"
	"net"

	"github.com/IBM/sarama"
	"github.com/gravitational/trace"

	"github.com/gravitational/teleport/lib/srv/db/common"
	"github.com/gravitational/teleport/lib/srv/db/mongodb/protocol"
	"github.com/gravitational/teleport/lib/utils"
)

const (
	KB = 1000
	MB = 1000 * KB
)

// NewEngine create new Kafka engine.
func NewEngine(ec common.EngineConfig) common.Engine {
	return &Engine{
		EngineConfig:   ec,
		maxMessageSize: 1 * MB,
	}
}

// Engine implements the Kafka database service that accepts client
// connections coming over reverse tunnel from the proxy and proxies
// them between the proxy and the MongoDB database instance.
//
// Implements common.Engine.
type Engine struct {
	// EngineConfig is the common database engine configuration.
	common.EngineConfig
	// clientConn is an incoming client connection.
	clientConn net.Conn
	// maxMessageSize is the max message size.
	maxMessageSize uint32
	// serverConnected specifies whether server connection has been created.
	serverConnected bool
}

// InitializeConnection initializes the client connection.
func (e *Engine) InitializeConnection(clientConn net.Conn, _ *common.Session) error {
	e.clientConn = clientConn
	return nil
}

// SendError sends an error to the connected client in MongoDB understandable format.
func (e *Engine) SendError(err error) {
	if err != nil && !utils.IsOKNetworkError(err) {
		e.replyError(e.clientConn, nil, err)
	}
}

// replyError sends the error to client. It is currently assumed that this
// function will only be called when HandleConnection terminates.
func (e *Engine) replyError(clientConn net.Conn, replyTo protocol.Message, err error) {
	// noop
}

// HandleConnection processes the connection from MongoDB proxy coming
// over reverse tunnel.
//
// It handles all necessary startup actions, authorization and acts as a
// middleman between the proxy and the database intercepting and interpreting
// all messages i.e. doing protocol parsing.
func (e *Engine) HandleConnection(ctx context.Context, sessionCtx *common.Session) error {
	// Establish connection to the Kafka server.
	serverConn, closeFn, err := e.connect(ctx, sessionCtx)
	if err != nil {
		return trace.Wrap(err, "error connecting to the database")
	}
	defer closeFn()

	e.serverConnected = true

	// Start reading client messages and sending them to server.
	for {
		clientMessage, err := protocol.ReadMessage(e.clientConn, e.maxMessageSize)
		if err != nil {
			return trace.Wrap(err)
		}
		err = e.handleClientMessage(ctx, sessionCtx, clientMessage, e.clientConn, serverConn)
		if err != nil {
			return trace.Wrap(err)
		}
	}
}

func (e *Engine) handleClientMessage(ctx context.Context, sessionCtx *common.Session, clientMessage protocol.Message, clientConn net.Conn, serverConn sarama.ClusterAdmin) error {
	// FIXME: We dont care about authorization - whether the user can create a topic or not.

	// FIXME: For now, we only assume listtopics is the only supported arg

	// synchronous call, so this will block until the server replies.
	topics, err := serverConn.ListTopics()
	if err != nil {
		return trace.Wrap(err)
	}

	// serialize the topics
	serverMessage, err := json.Marshal(topics)
	if err != nil {
		return trace.Wrap(err)
	}

	// ... and pass it back to the client.
	_, err = clientConn.Write(serverMessage)
	if err != nil {
		return trace.Wrap(err)
	}
	return nil
}
