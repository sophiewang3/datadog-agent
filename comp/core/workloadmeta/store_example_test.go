// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2022-present Datadog, Inc.

package workloadmeta

import (
	"fmt"
	"testing"

	"github.com/DataDog/datadog-agent/comp/core/config"
	"github.com/DataDog/datadog-agent/comp/core/log"
	"github.com/DataDog/datadog-agent/pkg/util/fxutil"
	"go.uber.org/fx"
)

func TestExampleStoreSubscribe(t *testing.T) {

	deps := fxutil.Test[dependencies](t, fx.Options(
		log.MockModule,
		config.MockModule,
		fx.Supply(NewParams()),
	))

	s := newWorkloadMeta(deps).(*workloadmeta)
	SetGlobalStore(s)

	filterParams := FilterParams{
		Kinds:     []Kind{KindContainer},
		Source:    SourceRuntime,
		EventType: EventTypeAll,
	}
	filter := NewFilter(&filterParams)

	ch := GetGlobalStore().Subscribe("test", NormalPriority, filter)

	go func() {
		for bundle := range ch {
			// close Ch to indicate that the Store can proceed to the next subscriber
			close(bundle.Ch)

			for _, evt := range bundle.Events {
				if evt.Type == EventTypeSet {
					fmt.Printf("new/updated container: %s\n", evt.Entity.GetID().ID)
				} else {
					fmt.Printf("container removed: %s\n", evt.Entity.GetID().ID)
				}
			}
		}
	}()

	// unsubscribe immediately so that this example does not hang
	GetGlobalStore().Unsubscribe(ch)

	// Output:
}
