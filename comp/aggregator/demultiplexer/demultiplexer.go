// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2023-present Datadog, Inc.

package demultiplexer

import (
	"context"

	"github.com/DataDog/datadog-agent/comp/aggregator/diagnosesendermanager"
	"github.com/DataDog/datadog-agent/comp/core/log"
	"github.com/DataDog/datadog-agent/comp/forwarder/defaultforwarder"
	"github.com/DataDog/datadog-agent/pkg/aggregator"
	"github.com/DataDog/datadog-agent/pkg/aggregator/sender"
	"github.com/DataDog/datadog-agent/pkg/util/hostname"
	"go.uber.org/fx"
)

type dependencies struct {
	fx.In
	Log             log.Component
	SharedForwarder defaultforwarder.Component

	Params Params
}

type demultiplexer struct {
	*aggregator.AgentDemultiplexer
}

type provides struct {
	fx.Out
	Comp Component

	// Both demultiplexer.Component and diagnosesendermanager.Component expose a different instance of SenderManager.
	// It means that diagnosesendermanager.Component must not be used when there is demultiplexer.Component instance.
	//
	// newDemultiplexer returns both demultiplexer.Component and diagnosesendermanager.Component (Note: demultiplexer.Component
	// implements diagnosesendermanager.Component). This has the nice consequence of preventing having
	// demultiplexer.Module and diagnosesendermanagerimpl.Module in the same fx.App because there would
	// be two ways to create diagnosesendermanager.Component.
	SenderManager diagnosesendermanager.Component
}

func newDemultiplexer(deps dependencies) (provides, error) {
	hostnameDetected, err := hostname.Get(context.TODO())
	if err != nil {
		if deps.Params.ContinueOnMissingHostname {
			deps.Log.Warnf("Error getting hostname: %s", err)
			hostnameDetected = ""
		} else {
			return provides{}, deps.Log.Errorf("Error while getting hostname, exiting: %v", err)
		}
	}

	agentDemultiplexer := aggregator.InitAndStartAgentDemultiplexer(
		deps.Log,
		deps.SharedForwarder,
		deps.Params.Options,
		hostnameDetected)
	demultiplexer := demultiplexer{
		AgentDemultiplexer: agentDemultiplexer,
	}

	return provides{
		Comp:          demultiplexer,
		SenderManager: demultiplexer,
	}, nil
}

// LazyGetSenderManager gets an instance of SenderManager lazily.
func (demux demultiplexer) LazyGetSenderManager() (sender.SenderManager, error) {
	return demux, nil
}
