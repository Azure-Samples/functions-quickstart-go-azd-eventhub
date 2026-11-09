package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/azure/azure-functions-golang-worker/sdk"
	"github.com/azure/azure-functions-golang-worker/sdk/bindings"
	"github.com/azure/azure-functions-golang-worker/worker"
)

type Reading struct {
	DeviceID    string  `json:"deviceId"`
	Temperature float64 `json:"temperature"`
}

func main() {
	worker.Start(newFunctionApp())
}

func newFunctionApp() *sdk.App {
	app := sdk.FunctionApp()
	delay := 5 * time.Second
	app.EventHub("processTelemetry", processTelemetry,
		sdk.WithEventHubName("telemetry"),
		sdk.WithConsumerGroup("functions"),
		sdk.WithConnection("EventHubConnection"),
		sdk.WithRetry(&sdk.RetryOptions{
			Strategy: sdk.FixedDelay, MaxRetryCount: 3, DelayInterval: &delay,
		}),
	)
	return app
}

func processTelemetry(ctx context.Context, event bindings.EventHubMessage) error {
	var reading Reading
	if err := json.Unmarshal(event.Body, &reading); err != nil {
		return fmt.Errorf("decode reading: %w", err)
	}
	if reading.DeviceID == "" {
		return fmt.Errorf("deviceId is required")
	}
	slog.InfoContext(ctx, "telemetry processed",
		"device_id", reading.DeviceID,
		"temperature", reading.Temperature,
		"sequence_number", event.SequenceNumber,
		"offset", event.Offset,
	)
	return nil
}
