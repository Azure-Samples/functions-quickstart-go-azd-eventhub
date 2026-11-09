param name string
param location string = resourceGroup().location
param tags object = {}
param principalId string
param userPrincipalId string = ''

resource namespace 'Microsoft.EventHub/namespaces@2024-01-01' = {
  name: name
  location: location
  tags: tags
  sku: {
    name: 'Standard'
    tier: 'Standard'
    capacity: 1
  }
  properties: {
    minimumTlsVersion: '1.2'
    disableLocalAuth: true
    publicNetworkAccess: 'Enabled'
  }
}

resource hub 'Microsoft.EventHub/namespaces/eventhubs@2024-01-01' = {
  parent: namespace
  name: 'telemetry'
  properties: {
    partitionCount: 2
    messageRetentionInDays: 1
  }
}

resource consumerGroup 'Microsoft.EventHub/namespaces/eventhubs/consumergroups@2024-01-01' = {
  parent: hub
  name: 'functions'
}

var receiverRole = 'a638d3c7-ab3a-418d-83e6-5f17a39d4fde'
var senderRole = '2b629674-e913-4c01-ae53-ef4638d8f975'

resource receiver 'Microsoft.Authorization/roleAssignments@2022-04-01' = {
  name: guid(hub.id, principalId, receiverRole)
  scope: hub
  properties: {
    principalId: principalId
    principalType: 'ServicePrincipal'
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', receiverRole)
  }
}

resource sender 'Microsoft.Authorization/roleAssignments@2022-04-01' = if (!empty(userPrincipalId)) {
  name: guid(hub.id, userPrincipalId, senderRole)
  scope: hub
  properties: {
    principalId: userPrincipalId
    principalType: 'User'
    roleDefinitionId: subscriptionResourceId('Microsoft.Authorization/roleDefinitions', senderRole)
  }
}

output fullyQualifiedNamespace string = '${namespace.name}.servicebus.windows.net'
