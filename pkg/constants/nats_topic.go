package constants

import "fmt"

const (
	BrokerRpcWildcardTopic              = "brokerRpc.>"
	BrokerRpcGetCometConfigurationTopic = "brokerRpc.getCometConfiguration"
	BrokerRpcVerifyAuthTokenTopic       = "brokerRpc.verifyAuthToken"

	UpgoingMessageTopic = "broker.upgoingMessage"
	EventTopic          = "broker.event"
)

const (
	cometRpcGetConnectionInfo = "cometRpc.%s.getConnectionInfo"

	cometDowngoingMessageTopic = "comet.%s.downgoingMessage"
)

func CometRpcGetConnectionInfoTopic(machineId string) string {
	return fmt.Sprintf(cometRpcGetConnectionInfo, machineId)
}

func CometDowngoingMessageTopic(machineId string) string {
	return fmt.Sprintf(cometDowngoingMessageTopic, machineId)
}

const (
	BrokerGroup = "brokerGroup"
)
