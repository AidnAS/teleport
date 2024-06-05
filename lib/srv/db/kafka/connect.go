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
	"crypto/tls"

	"github.com/IBM/sarama"
	"github.com/gravitational/trace"

	"github.com/gravitational/teleport/lib/srv/db/common"
)

// connect returns connection to a kafka server.
//
// When connecting to a replica set, returns connection to the server selected
// based on the read preference connection string option. This allows users to
// configure database access to always connect to a secondary for example.
func (e *Engine) connect(ctx context.Context, sessionCtx *common.Session) (sarama.ClusterAdmin, func(), error) {
	tlsConfig, err := e.Auth.GetTLSConfig(ctx, sessionCtx)
	if err != nil {
		return nil, nil, trace.Wrap(err)
	}

	config := sarama.NewConfig()
	config.Net.TLS.Enable = true
	config.Net.TLS.Config = tlsConfig
	config.Net.TLS.Config.InsecureSkipVerify = false    // TODO: get the insecure flag from the database
	config.Net.TLS.Config.ClientAuth = tls.NoClientCert // FIXME: get the client cert from the database
	// config.Net.TLS.Config.RootCAs = args.TLS.RootCAs // guess we might need this to verify server certs

	client, err := sarama.NewClusterAdmin([]string{sessionCtx.Database.GetURI()}, config)

	/* FIXME: this is the confluent-kafka-go client
	// create kafka admin client
	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{
		// https://github.com/confluentinc/librdkafka/blob/master/CONFIGURATION.md
		"bootstrap.servers":   sessionCtx.Database.GetURI(),
		"security.protocol":   "ssl",
		"ssl.key.pem":         sessionCtx.Database.GetCA(), // TODO: get the teleport client cert
		"ssl.certificate.pem": tls.CACert,                  // TODO: get the Client's public key string (PEM format) used for authentication.
		"ssl.ca.pem":          sessionCtx.Database.GetCA(), // CA certificate string (PEM format) for verifying the broker's key.
	})
	*/

	closeFn := func() {
		client.Close()
	}
	return client, closeFn, err
}
