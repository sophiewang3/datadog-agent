// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package statsd

import (
	"net"
	"os"
	"strconv"

	"github.com/DataDog/datadog-go/v5/statsd"
)

// Client is a global Statsd client. When a client is configured via Configure,
// that becomes the new global Statsd client in the package.
var Client *statsd.Client

// Configure creates a statsd client from a dogweb.ini style config file and set it to the global Statsd.
func Configure(host string, port int, lookInEnv bool) error {
	var statsdAddr string
	if lookInEnv {
		statsdAddr = os.Getenv("STATSD_URL")
	}

	if statsdAddr == "" {
		statsdAddr = net.JoinHostPort(host, strconv.Itoa(port))
	}

	options := []statsd.Option{
		// Create a separate client for the telemetry to be sure we don't loose it.
		statsd.WithTelemetryAddr(statsdAddr),
		// Enable recommended settings to reduce the number of packets sent and reduce
		// potential lock contention on the critical path.
		statsd.WithChannelMode(),
		statsd.WithClientSideAggregation(),
		statsd.WithExtendedClientSideAggregation(),
	}
	client, err := statsd.New(statsdAddr, options...)
	if err != nil {
		return err
	}

	Client = client
	return nil
}
