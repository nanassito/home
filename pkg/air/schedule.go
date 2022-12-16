package air

import (
	"fmt"
	"time"
)

func (s *Server) ApplySchedules() {
	tz, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		logger.Fatalf("Can't load timezone %v", err)
	}
	timenow := time.Now().In(tz)
	now := fmt.Sprintf("%02d:%02d", timenow.Hour(), timenow.Minute())
	for name, weekSchedule := range s.State.Schedules {
		if !weekSchedule.IsActive {
			continue
		}
		schedule := weekSchedule.Weekday
		if time.Now().Weekday() == time.Saturday || time.Now().Weekday() == time.Sunday {
			schedule = weekSchedule.Weekend
		}
		for _, window := range schedule {
			if window.Start <= now && now < window.End {
				room := s.State.Rooms[weekSchedule.RoomName]
				if room.DesiredTemperatureRange != window.Settings {
					logger.Printf("Schedule %s changed Room %s to %v\n", name, room.Name, window.Settings)
					room.DesiredTemperatureRange = window.Settings
				}
			}
		}
	}
}
