package switches_test

import (
	"context"
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/nanassito/home/pkg/switches"
	"github.com/nanassito/home/pkg/switches_proto"
)

func TestActivate(t *testing.T) {
	is := is.New(t)
	server := switches.Server{
		State: &switches.ServerState{
			SwitchIDs: []string{"testValve"},
			BySwitchID: map[string]*switches.State{
				"testValve": {
					SwitchID:       "testValve",
					Config:         &switches_proto.SwitchConfig{},
					Requests:       []switches.Request{},
					ReportedActive: false,
				},
			},
		},
	}
	before := time.Now()
	resp, err := server.Activate(context.Background(), &switches_proto.ReqActivate{
		SwitchID:        "testValve",
		DurationSeconds: int64(10),
		ClientID:        "unitTests",
	})
	after := time.Now()
	is.NoErr(err)
	is.True(before.Add(9 * time.Second).Before(time.Unix(resp.GetActiveUntil(), 0)))
	is.True(after.Add(11 * time.Second).After(time.Unix(resp.GetActiveUntil(), 0)))
}

func TestPersistence(t *testing.T) {
	is := is.New(t)
	server := switches.Server{
		State: &switches.ServerState{
			SwitchIDs: []string{"testValve"},
			BySwitchID: map[string]*switches.State{
				"testValve": {
					SwitchID:       "testValve",
					Config:         &switches_proto.SwitchConfig{},
					Requests:       []switches.Request{},
					ReportedActive: false,
				},
			},
		},
	}
	_, err := server.Activate(context.Background(), &switches_proto.ReqActivate{
		SwitchID:        "testValve",
		DurationSeconds: int64(10),
		ClientID:        "unitTests",
	})
	is.NoErr(err)
	resp, err := server.Status(context.Background(), &switches_proto.ReqStatus{
		SwitchID: "testValve",
	})
	is.NoErr(err)
	is.Equal(1, len(resp.Requests))
}

func TestStatus(t *testing.T) {
	is := is.New(t)
	server := switches.Server{
		State: &switches.ServerState{
			SwitchIDs: []string{"testValve"},
			BySwitchID: map[string]*switches.State{
				"testActiveValve": {
					SwitchID: "testActiveValve",
					Config:   &switches_proto.SwitchConfig{},
					Requests: []switches.Request{
						{
							ClientID: "unitTest",
							From:     time.Now().Add(-20 * time.Second),
							Until:    time.Now().Add(-10 * time.Second),
						},
						{
							ClientID: "unitTest",
							From:     time.Now().Add(-10 * time.Second),
							Until:    time.Now().Add(10 * time.Second),
						},
					},
				},
				"testInactiveValve": {
					SwitchID: "testInactiveValve",
					Config:   &switches_proto.SwitchConfig{},
					Requests: []switches.Request{
						{
							ClientID: "unitTest",
							From:     time.Now().Add(-20 * time.Second),
							Until:    time.Now().Add(-10 * time.Second),
						},
					},
				},
			},
		},
	}

	resp, err := server.Status(context.Background(), &switches_proto.ReqStatus{
		SwitchID: "testActiveValve",
	})
	is.NoErr(err)
	is.True(resp.IsActive)

	resp, err = server.Status(context.Background(), &switches_proto.ReqStatus{
		SwitchID: "testInactiveValve",
	})
	is.NoErr(err)
	is.True(!resp.IsActive)
}
