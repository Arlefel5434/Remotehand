package main

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestQRNetworkTransitions(t *testing.T) {
	home := phoneLink{"Ethernet", "192.168.1.58"}
	hotspot := phoneLink{"Wi-Fi", "172.20.10.2"}
	scans := [][]phoneLink{nil, nil, {home}, {home}, nil, {hotspot}, {home, hotspot}}
	ticks := make(chan time.Time, len(scans)-1)
	for range len(scans) - 1 {
		ticks <- time.Time{}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var shown [][]phoneLink
	monitorLinks(ctx, ticks, func() []phoneLink {
		current := scans[0]
		scans = scans[1:]
		if len(scans) == 0 {
			cancel()
		}
		return current
	}, func(links []phoneLink) { shown = append(shown, links) })
	want := [][]phoneLink{nil, {home}, nil, {hotspot}, {home}}
	if !reflect.DeepEqual(shown, want) {
		t.Fatalf("displayed %v; want %v", shown, want)
	}
}
