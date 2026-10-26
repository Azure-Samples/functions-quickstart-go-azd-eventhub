package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs"
)

func main() {
	if err := send(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func send() error {
	namespace := os.Getenv("EVENTHUB_NAMESPACE")
	if namespace == "" {
		return fmt.Errorf("set EVENTHUB_NAMESPACE to the fully qualified namespace")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}
	client, err := azeventhubs.NewProducerClient(namespace, "telemetry", credential, nil)
	if err != nil {
		return err
	}
	defer client.Close(context.Background())
	batch, err := client.NewEventDataBatch(ctx, nil)
	if err != nil {
		return err
	}
	if err := batch.AddEventData(&azeventhubs.EventData{
		Body: []byte(`{"deviceId":"sensor-001","temperature":23.5}`),
	}, nil); err != nil {
		return err
	}
	if err := client.SendEventDataBatch(ctx, batch, nil); err != nil {
		return err
	}
	fmt.Println("Sent reading for sensor-001 to telemetry.")
	return nil
}
