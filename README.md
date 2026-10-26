<!--
---
name: Azure Functions Go Event Hubs Trigger using Azure Developer CLI
description: Process device telemetry from Event Hubs in Go on Azure Functions Flex Consumption using managed identity.
page_type: sample
languages:
- azdeveloper
- go
products:
- azure
- azure-functions
- azure-event-hubs
urlFragment: functions-quickstart-go-azd-eventhub
---
-->

# Azure Functions Go Event Hubs Trigger

This sample follows the [Go timer quickstart](https://github.com/Azure-Samples/functions-quickstart-go-azd-timer)
and deploys to Linux Flex Consumption with `azd`. Go worker packages are preview
packages. The `processTelemetry` function consumes JSON device readings from the
`telemetry` event hub using the dedicated `functions` consumer group.

```json
{"deviceId":"sensor-001","temperature":23.5}
```

The handler logs the device, temperature, sequence number, and offset.
Malformed readings return an error and are retried three times with a five-second
fixed delay. Event Hubs is an at-least-once stream, not a queue: there is no
dead-letter queue, and after retries are exhausted processing can advance past
the failed event. Production apps should persist invalid events explicitly and
make side effects idempotent.

## Prerequisites

- [Go 1.25 or later](https://go.dev/dl/)
- [Azure Functions Core Tools 4.12 or later](https://learn.microsoft.com/azure/azure-functions/functions-run-local)
- [Azure CLI](https://learn.microsoft.com/cli/azure/install-azure-cli)
- [Azure Developer CLI](https://learn.microsoft.com/azure/developer/azure-developer-cli/install-azd)
- Azure subscription with permissions to provision resources and assign roles
- [Azurite](https://learn.microsoft.com/azure/storage/common/storage-use-azurite) for local host checkpoints
- Optional: VS Code with the Azure Functions and Go extensions

## Initialize

```shell
git clone https://github.com/Azure-Samples/functions-quickstart-go-azd-eventhub.git
cd functions-quickstart-go-azd-eventhub
```

## Deploy

Sign in and provision a sender role for your signed-in user:

```powershell
az login
azd auth login
azd env new eventhub-dev
azd env set AZURE_PRINCIPAL_ID (az ad signed-in-user show --query id -o tsv)
azd up
```

The infrastructure creates a Standard Event Hubs namespace (one throughput unit),
the event hub and consumer group, managed identity, host storage, Application
Insights, and a VNet with private endpoints for **host storage** by default.
The Event Hubs namespace has a public endpoint secured by Entra ID; SAS
authentication is disabled. Standard Event Hubs has ongoing charges.

The app identity has **Azure Event Hubs Data Receiver** scoped to the hub; the
optional test user has **Azure Event Hubs Data Sender**. Event Hubs checkpoints
use the identity-based host storage connection.

To disable the host storage VNet for development:

```shell
azd env set VNET_ENABLED false
azd up
```

## Send an event and verify

The included Go publisher uses `DefaultAzureCredential` (including your Azure CLI
login), not an Event Hubs connection string:

```powershell
$deployment = azd env get-values --output json | ConvertFrom-Json
$env:EVENTHUB_NAMESPACE = $deployment.EVENTHUB_NAMESPACE
go run ./cmd/send
```

Allow a few minutes for initial RBAC propagation. In the function app's
Application Insights **Logs**, run:

```kusto
traces
| where timestamp > ago(15m)
| where message contains "telemetry processed"
| project timestamp, message, customDimensions
```

You should see `device_id=sensor-001` and `temperature=23.5`.

## Run locally

Use an Azure Event Hubs development namespace and an identity with Data Receiver
on `telemetry`. The sample's sender role alone does not authorize a local listener;
assign Data Receiver to your development identity separately.

```powershell
Copy-Item local.settings.example.json local.settings.json
```

Replace `EventHubConnection__fullyQualifiedNamespace` with your namespace
hostname. Start Azurite in another terminal and then run `func start`.
The host builds the app automatically. No `function.json` is required.
Do not use the same consumer group for local and deployed listeners concurrently:
stop one listener or provision a separate development consumer group.
The sample uses real Azure Event Hubs; Azurite emulates only checkpoint storage.

## Tests, redeployment, and cleanup

```shell
go test ./...
go vet ./...
azd up
azd down
```

Tests cover valid readings (including zero temperature), malformed input, and
the trigger's connection, consumer group, and retry configuration.
