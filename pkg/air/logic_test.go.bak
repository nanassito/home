package air_test

import (
	"testing"
	"time"

	"github.com/matryer/is"
	"github.com/nanassito/home/pkg/air"
	"github.com/nanassito/home/pkg/air_proto"
)

func TestInferGeneralMode(t *testing.T) {

	getServer := func(mode air_proto.Hvac_Mode) air.Server {
		return air.Server{
			State: &air_proto.ServerState{
				Rooms: map[string]*air_proto.RoomState{
					"A": {
						Hvacs: map[string]*air_proto.Hvac{
							"a": {Control: air_proto.Hvac_CONTROL_ROOM},
							"b": {
								HvacName: "b",
								Control:  air_proto.Hvac_CONTROL_HVAC,
								ReportedState: &air_proto.Hvac_State{
									Mode: mode,
								},
							},
						},
						Sensor: &air_proto.Sensor{
							// Location:         "",
							Temperature:      0,
							LastReportedAtTs: time.Now().Unix(),
						},
					},
					"C": {
						Hvacs: map[string]*air_proto.Hvac{
							"c": {Control: air_proto.Hvac_CONTROL_ROOM},
						},
						Sensor: &air_proto.Sensor{
							// Location:         "",
							Temperature:      0,
							LastReportedAtTs: time.Now().Unix(),
						},
					},
				},
			},
		}
	}

	t.Run("defers to explicitly controled hvac", func(t *testing.T) {
		is := is.New(t)
		server := getServer(air_proto.Hvac_MODE_DRY)
		is.Equal(air_proto.Hvac_MODE_DRY, server.InferGeneralMode())
	})

	t.Run("ignore explicitly controled hvac that is off", func(t *testing.T) {
		is := is.New(t)
		server := getServer(air_proto.Hvac_MODE_OFF)
		is.Equal(air_proto.Hvac_MODE_HEAT, server.InferGeneralMode())
	})
}

func TestDecideHeatUp(t *testing.T) {
	getServer := func() air.Server {
		return air.Server{
			State: &air_proto.ServerState{
				Rooms: map[string]*air_proto.RoomState{
					"A": {
						RoomName: "A",
						Hvacs: map[string]*air_proto.Hvac{
							"a": {
								HvacName:     "a",
								Control:      air_proto.Hvac_CONTROL_ROOM,
								DesiredState: &air_proto.Hvac_State{},
							},
						},
						Sensor: &air_proto.Sensor{
							Temperature:      19,
							LastReportedAtTs: time.Now().Unix(),
						},
						DesiredTemperatureRange: &air_proto.TemperatureRange{Min: 20},
					},
				},
				OutsideSensor: &air_proto.Sensor{
					Temperature:      15,
					LastReportedAtTs: time.Now().Unix(),
				},
			},
			GetLast30mHvacsTemperature:  func() (map[string]float64, error) { return map[string]float64{}, nil },
			Get30mRoomTemperatureDeltas: func() (map[string]float64, error) { return map[string]float64{}, nil },
		}
	}

	t.Run("The Sun will warm up the room", func(t *testing.T) {
		is := is.New(t)
		server := getServer()
		server.State.Rooms["A"].Sensor.Temperature = 21
		server.State.OutsideSensor.Temperature = 25
		server.DecideHeatUp()
		is.Equal(server.State.Rooms["A"].Hvacs["a"].DesiredState.Mode, air_proto.Hvac_MODE_OFF)
	})
	t.Run("Room needs to be warmed up", func(t *testing.T) {
		is := is.New(t)
		server := getServer()
		server.State.Rooms["A"].Sensor.Temperature = 16
		server.DecideHeatUp()
		is.Equal(server.State.Rooms["A"].Hvacs["a"].DesiredState.Mode, air_proto.Hvac_MODE_HEAT)
	})
	t.Run("Room is warm but not enough to stop", func(t *testing.T) {
		is := is.New(t)
		server := getServer()
		server.State.Rooms["A"].Sensor.Temperature = 22
		server.DecideHeatUp()
		is.Equal(server.State.Rooms["A"].Hvacs["a"].DesiredState.Mode, air_proto.Hvac_MODE_HEAT)
	})
	t.Run("Room is warm enough to stop", func(t *testing.T) {
		is := is.New(t)
		server := getServer()
		server.State.Rooms["A"].Sensor.Temperature = 24
		server.DecideHeatUp()
		is.Equal(server.State.Rooms["A"].Hvacs["a"].DesiredState.Mode, air_proto.Hvac_MODE_OFF)
	})
	t.Run("Room hot enough not to care about temperature gradient", func(t *testing.T) {
		is := is.New(t)
		server := getServer()
		server.State.Rooms["A"].Sensor.Temperature = 21
		server.DecideHeatUp()
		is.Equal(server.State.Rooms["A"].Hvacs["a"].DesiredState.Fan, air_proto.Hvac_FAN_AUTO)
	})
	t.Run("High temperature gradient", func(t *testing.T) {
		is := is.New(t)
		server := getServer()
		server.State.Rooms["A"].Sensor.Temperature = 19
		server.GetLast30mHvacsTemperature = func() (map[string]float64, error) {
			return map[string]float64{"a": 23}, nil
		}
		server.DecideHeatUp()
		is.Equal(server.State.Rooms["A"].Hvacs["a"].DesiredState.Fan, air_proto.Hvac_FAN_HIGH)
	})
	t.Run("Medium temperature gradient", func(t *testing.T) {
		is := is.New(t)
		server := getServer()
		server.State.Rooms["A"].Sensor.Temperature = 19
		server.GetLast30mHvacsTemperature = func() (map[string]float64, error) {
			return map[string]float64{"a": 21}, nil
		}
		server.DecideHeatUp()
		is.Equal(server.State.Rooms["A"].Hvacs["a"].DesiredState.Fan, air_proto.Hvac_FAN_MEDIUM)
	})
	t.Run("No temperature gradient", func(t *testing.T) {
		is := is.New(t)
		server := getServer()
		server.State.Rooms["A"].Sensor.Temperature = 19
		server.GetLast30mHvacsTemperature = func() (map[string]float64, error) {
			return map[string]float64{"a": 20}, nil
		}
		server.DecideHeatUp()
		// Then we put the fan in auto mode.
		is.Equal(server.State.Rooms["A"].Hvacs["a"].DesiredState.Fan, air_proto.Hvac_FAN_AUTO)
	})
	t.Run("We don't have historical data for the hvac", func(t *testing.T) {
		is := is.New(t)
		server := getServer()
		server.GetLast30mHvacsTemperature = func() (map[string]float64, error) {
			return map[string]float64{}, nil
		}
		server.DecideHeatUp()
		// Then we put the fan in auto mode.
		is.Equal(server.State.Rooms["A"].Hvacs["a"].DesiredState.Fan, air_proto.Hvac_FAN_AUTO)
	})
	t.Run("We don't have historical data for the room", func(t *testing.T) {
		is := is.New(t)
		server := getServer()
		server.Get30mRoomTemperatureDeltas = func() (map[string]float64, error) {
			return map[string]float64{}, nil
		}
		server.State.Rooms["A"].Hvacs["a"].TemperatureOffset = 2
		server.DecideHeatUp()
		// Then we leave the offset untouched.
		is.Equal(server.State.Rooms["A"].Hvacs["a"].TemperatureOffset, int64(2))
	})
	t.Run("Temperature is rising and it's already too hot", func(t *testing.T) {
		is := is.New(t)
		server := getServer()
		server.State.Rooms["A"].Sensor.Temperature = 22
		server.Get30mRoomTemperatureDeltas = func() (map[string]float64, error) {
			return map[string]float64{"A": 1}, nil
		}
		server.State.Rooms["A"].Hvacs["a"].TemperatureOffset = 0
		server.DecideHeatUp()
		// Then lower the target temperature
		is.Equal(server.State.Rooms["A"].Hvacs["a"].TemperatureOffset, int64(-1))
	})
	t.Run("Temperature is dropping and it's already too cold", func(t *testing.T) {
		is := is.New(t)
		server := getServer()
		server.State.Rooms["A"].Sensor.Temperature = 19
		server.Get30mRoomTemperatureDeltas = func() (map[string]float64, error) {
			return map[string]float64{"A": -1}, nil
		}
		server.State.Rooms["A"].Hvacs["a"].TemperatureOffset = 0
		server.DecideHeatUp()
		// Then raise the target temperature
		is.Equal(server.State.Rooms["A"].Hvacs["a"].TemperatureOffset, int64(1))
	})
	t.Run("Temperature is dropping but it's too hot", func(t *testing.T) {
		is := is.New(t)
		server := getServer()
		server.State.Rooms["A"].Sensor.Temperature = 22
		server.Get30mRoomTemperatureDeltas = func() (map[string]float64, error) {
			return map[string]float64{"A": -1}, nil
		}
		server.State.Rooms["A"].Hvacs["a"].TemperatureOffset = 1
		server.DecideHeatUp()
		// Then keep the target temperature
		is.Equal(server.State.Rooms["A"].Hvacs["a"].TemperatureOffset, int64(1))
	})
	t.Run("Temperature is rising but it's too cold", func(t *testing.T) {
		is := is.New(t)
		server := getServer()
		server.State.Rooms["A"].Sensor.Temperature = 19
		server.Get30mRoomTemperatureDeltas = func() (map[string]float64, error) {
			return map[string]float64{"A": 1}, nil
		}
		server.State.Rooms["A"].Hvacs["a"].TemperatureOffset = 1
		server.DecideHeatUp()
		// Then keep the target temperature
		is.Equal(server.State.Rooms["A"].Hvacs["a"].TemperatureOffset, int64(1))
	})
}
