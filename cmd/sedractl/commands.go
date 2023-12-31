package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sedracoin/sedrad/infrastructure/network/netadapter/server/grpcserver/protowire"
)

var commandTypes = []reflect.Type{
	reflect.TypeOf(protowire.SedradMessage_AddPeerRequest{}),
	reflect.TypeOf(protowire.SedradMessage_GetConnectedPeerInfoRequest{}),
	reflect.TypeOf(protowire.SedradMessage_GetPeerAddressesRequest{}),
	reflect.TypeOf(protowire.SedradMessage_GetCurrentNetworkRequest{}),
	reflect.TypeOf(protowire.SedradMessage_GetInfoRequest{}),

	reflect.TypeOf(protowire.SedradMessage_GetBlockRequest{}),
	reflect.TypeOf(protowire.SedradMessage_GetBlocksRequest{}),
	reflect.TypeOf(protowire.SedradMessage_GetHeadersRequest{}),
	reflect.TypeOf(protowire.SedradMessage_GetBlockCountRequest{}),
	reflect.TypeOf(protowire.SedradMessage_GetBlockDagInfoRequest{}),
	reflect.TypeOf(protowire.SedradMessage_GetSelectedTipHashRequest{}),
	reflect.TypeOf(protowire.SedradMessage_GetVirtualSelectedParentBlueScoreRequest{}),
	reflect.TypeOf(protowire.SedradMessage_GetVirtualSelectedParentChainFromBlockRequest{}),
	reflect.TypeOf(protowire.SedradMessage_ResolveFinalityConflictRequest{}),
	reflect.TypeOf(protowire.SedradMessage_EstimateNetworkHashesPerSecondRequest{}),

	reflect.TypeOf(protowire.SedradMessage_GetBlockTemplateRequest{}),
	reflect.TypeOf(protowire.SedradMessage_SubmitBlockRequest{}),

	reflect.TypeOf(protowire.SedradMessage_GetMempoolEntryRequest{}),
	reflect.TypeOf(protowire.SedradMessage_GetMempoolEntriesRequest{}),
	reflect.TypeOf(protowire.SedradMessage_GetMempoolEntriesByAddressesRequest{}),

	reflect.TypeOf(protowire.SedradMessage_SubmitTransactionRequest{}),

	reflect.TypeOf(protowire.SedradMessage_GetUtxosByAddressesRequest{}),
	reflect.TypeOf(protowire.SedradMessage_GetBalanceByAddressRequest{}),
	reflect.TypeOf(protowire.SedradMessage_GetCoinSupplyRequest{}),

	reflect.TypeOf(protowire.SedradMessage_BanRequest{}),
	reflect.TypeOf(protowire.SedradMessage_UnbanRequest{}),
}

type commandDescription struct {
	name       string
	parameters []*parameterDescription
	typeof     reflect.Type
}

type parameterDescription struct {
	name   string
	typeof reflect.Type
}

func commandDescriptions() []*commandDescription {
	commandDescriptions := make([]*commandDescription, len(commandTypes))

	for i, commandTypeWrapped := range commandTypes {
		commandType := unwrapCommandType(commandTypeWrapped)

		name := strings.TrimSuffix(commandType.Name(), "RequestMessage")
		numFields := commandType.NumField()

		var parameters []*parameterDescription
		for i := 0; i < numFields; i++ {
			field := commandType.Field(i)

			if !isFieldExported(field) {
				continue
			}

			parameters = append(parameters, &parameterDescription{
				name:   field.Name,
				typeof: field.Type,
			})
		}
		commandDescriptions[i] = &commandDescription{
			name:       name,
			parameters: parameters,
			typeof:     commandTypeWrapped,
		}
	}

	return commandDescriptions
}

func (cd *commandDescription) help() string {
	sb := &strings.Builder{}
	sb.WriteString(cd.name)
	for _, parameter := range cd.parameters {
		_, _ = fmt.Fprintf(sb, " [%s]", parameter.name)
	}
	return sb.String()
}
