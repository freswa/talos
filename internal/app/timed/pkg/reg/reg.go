// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package reg

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	timeapi "github.com/talos-systems/talos/api/time"
	"github.com/talos-systems/talos/internal/app/timed/pkg/ntp"
)

// Registrator is the concrete type that implements the factory.Registrator and
// timeapi.Init interfaces.
type Registrator struct {
	Timed *ntp.NTP
}

// NewRegistrator builds new Registrator instance
func NewRegistrator(n *ntp.NTP) *Registrator {
	return &Registrator{
		Timed: n,
	}
}

// Register implements the factory.Registrator interface.
func (r *Registrator) Register(s *grpc.Server) {
	timeapi.RegisterTimeServiceServer(s, r)
}

// Time issues a query to the configured ntp server and displays the results
func (r *Registrator) Time(ctx context.Context, in *empty.Empty) (reply *timeapi.TimeResponse, err error) {
	reply = &timeapi.TimeResponse{}

	rt, err := r.Timed.Query()
	if err != nil {
		return reply, err
	}

	return genProtobufTimeResponse(r.Timed.GetTime(), rt.Time, r.Timed.Server)
}

// TimeCheck issues a query to the specified ntp server and displays the results
func (r *Registrator) TimeCheck(ctx context.Context, in *timeapi.TimeRequest) (reply *timeapi.TimeResponse, err error) {
	reply = &timeapi.TimeResponse{}

	tc, err := ntp.NewNTPClient(ntp.WithServer(in.Server))
	if err != nil {
		return reply, err
	}

	rt, err := tc.Query()
	if err != nil {
		return reply, err
	}

	return genProtobufTimeResponse(tc.GetTime(), rt.Time, in.Server)
}

func genProtobufTimeResponse(local, remote time.Time, server string) (*timeapi.TimeResponse, error) {
	resp := &timeapi.TimeResponse{}

	localpbts, err := ptypes.TimestampProto(local)
	if err != nil {
		return resp, err
	}

	remotepbts, err := ptypes.TimestampProto(remote)
	if err != nil {
		return resp, err
	}

	resp = &timeapi.TimeResponse{
		Messages: []*timeapi.Time{
			{
				Server:     server,
				Localtime:  localpbts,
				Remotetime: remotepbts,
			},
		},
	}

	return resp, nil
}
