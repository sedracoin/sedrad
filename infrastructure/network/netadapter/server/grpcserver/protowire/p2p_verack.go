package protowire

import (
	"github.com/sedracoin/sedrad/app/appmessage"
	"github.com/pkg/errors"
)

func (x *SedradMessage_Verack) toAppMessage() (appmessage.Message, error) {
	if x == nil {
		return nil, errors.Wrapf(errorNil, "SedradMessage_Verack is nil")
	}
	return &appmessage.MsgVerAck{}, nil
}

func (x *SedradMessage_Verack) fromAppMessage(_ *appmessage.MsgVerAck) error {
	return nil
}
