// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// Package sysprobeconfig implements a component to handle system-probe configuration.  This
// component temporarily wraps pkg/config.
//
// This component initializes pkg/config based on the bundle params, and
// will return the same results as that package.  This is to support migration
// to a component architecture.  When no code still uses pkg/config, that
// package will be removed.
//
// The mock component does nothing at startup, beginning with an empty config.
// It also overwrites the pkg/config.SystemProbe for the duration of the test.
package sysprobeconfig

import (
	sysconfig "github.com/DataDog/datadog-agent/cmd/system-probe/config"
	"github.com/DataDog/datadog-agent/pkg/config"
)

// team: ebpf-platform

// Component is the component type.
type Component interface {
	config.Reader

	// Warnings returns config warnings collected during setup.
	Warnings() *config.Warnings

	// SysProbeObject returns the wrapper sysconfig
	SysProbeObject() *sysconfig.Config
}
