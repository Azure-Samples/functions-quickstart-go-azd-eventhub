package main

import (
	"context"
	"testing"

	"github.com/azure/azure-functions-golang-worker/sdk"
	"github.com/azure/azure-functions-golang-worker/sdk/bindings"
)

func TestProcessTelemetry(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		wantError  bool
	}{
		{"reading", `{"deviceId":"sensor-001","temperature":23.5}`, false},
		{"zero temperature", `{"deviceId":"sensor-001","temperature":0}`, false},
		{"missing device", `{"temperature":23.5}`, true},
		{"invalid JSON", `{`, true},
		{"null", `null`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := processTelemetry(context.Background(), bindings.EventHubMessage{Body: []byte(tc.body)})
			if (err != nil) != tc.wantError {
				t.Fatalf("error = %v, wantError = %v", err, tc.wantError)
			}
		})
	}
}

func TestEventHubRegistration(t *testing.T) {
	count := 0
	newFunctionApp().GetRegisteredFunctions().Range(func(_, value any) bool {
		count++
		fn := value.(*sdk.RegisteredFunction)
		binding := fn.TriggerBinding()
		if fn.FuncName != "processTelemetry" || binding == nil || binding.EventHubBinding == nil {
			t.Fatal("missing Event Hubs trigger")
		}
		b := binding.EventHubBinding
		if b.EventHubName != "telemetry" || b.ConsumerGroup != "functions" || b.Connection != "EventHubConnection" {
			t.Fatalf("unexpected trigger configuration: %+v", b)
		}
		if fn.Retry == nil || fn.Retry.MaxRetryCount != 3 {
			t.Fatal("missing bounded retry policy")
		}
		return true
	})
	if count != 1 {
		t.Fatalf("registered functions = %d, want 1", count)
	}
}
