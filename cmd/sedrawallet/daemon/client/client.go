package client

import (
	"context"
	"github.com/sedracoin/sedrad/cmd/sedrawallet/daemon/server"
	"time"

	"github.com/pkg/errors"

	"github.com/sedracoin/sedrad/cmd/sedrawallet/daemon/pb"
	"google.golang.org/grpc"
)

// Connect connects to the sedrawalletd server, and returns the client instance
func Connect(address string) (pb.SedrawalletdClient, func(), error) {
	// Connection is local, so 1 second timeout is sufficient
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(server.MaxDaemonSendMsgSize)))
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, nil, errors.New("sedrawallet daemon is not running, start it with `sedrawallet start-daemon`")
		}
		return nil, nil, err
	}

	return pb.NewSedrawalletdClient(conn), func() {
		conn.Close()
	}, nil
}
