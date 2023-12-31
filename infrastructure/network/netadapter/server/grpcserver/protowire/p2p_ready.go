package protowire

import (
	"github.com/sedracoin/sedrad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *SedradMessage_Ready) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "SedradMessage_Ready is nil")
	}
	return &appmessage.MsgReady{}, nil
}

func (x *SedradMessage_Ready) fromAppMessage(_ *appmessage.MsgReady) error {
	return nil
}
