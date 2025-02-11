// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package packets

import (
	"sync"
	"time"
)

// Buffer is a buffer of packets that will automatically flush to the given
// output channel when it is full or after a configurable duration.
type Buffer struct {
	listenerID    string
	packets       Packets
	flushTimer    *time.Ticker
	bufferSize    uint
	outputChannel chan Packets
	closeChannel  chan struct{}
	m             sync.Mutex
}

// NewBuffer creates a new buffer of packets of specified size
func NewBuffer(bufferSize uint, flushTimer time.Duration, outputChannel chan Packets, listenerID string) *Buffer {
	pb := &Buffer{
		listenerID:    listenerID,
		bufferSize:    bufferSize,
		flushTimer:    time.NewTicker(flushTimer),
		outputChannel: outputChannel,
		packets:       make(Packets, 0, bufferSize),
		closeChannel:  make(chan struct{}),
	}
	go pb.flushLoop()
	return pb
}

func (pb *Buffer) flushLoop() {
	for {
		select {
		case <-pb.flushTimer.C:
			pb.m.Lock()
			pb.flush()
			tlmBufferFlushedTimer.Inc(pb.listenerID)
			pb.m.Unlock()
		case <-pb.closeChannel:
			return
		}
	}
}

// Append appends a packet to the packet buffer and flushes if the buffer size is to be exceeded.
func (pb *Buffer) Append(packet *Packet) {
	pb.m.Lock()
	defer pb.m.Unlock()
	pb.packets = append(pb.packets, packet)

	tlmBufferSize.Set(float64(len(pb.packets)), pb.listenerID)

	if uint(len(pb.packets)) >= pb.bufferSize {
		pb.flush()
		tlmBufferFlushedFull.Inc(pb.listenerID)
	}
}

func (pb *Buffer) flush() {
	if len(pb.packets) > 0 {
		t1 := time.Now()
		pb.outputChannel <- pb.packets
		t2 := time.Now()
		tlmListenerChannel.Observe(float64(t2.Sub(t1).Nanoseconds()), pb.listenerID)

		pb.packets = make(Packets, 0, pb.bufferSize)
	}
	tlmBufferSize.Set(float64(len(pb.packets)), pb.listenerID)
	// FIXME: it's not per listener
	tlmChannelSize.Set(float64(len(pb.outputChannel)), pb.listenerID)
}

// Close closes the packet buffer
func (pb *Buffer) Close() {
	close(pb.closeChannel)
	if pb.listenerID != "" {
		tlmBufferSize.Delete(pb.listenerID)
		tlmChannelSize.Delete(pb.listenerID)
		tlmBufferFlushedFull.Delete(pb.listenerID)
		tlmBufferFlushedTimer.Delete(pb.listenerID)
	}
}
