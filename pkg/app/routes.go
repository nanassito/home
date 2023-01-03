package app

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/nanassito/home/pkg/air_proto"
	"github.com/nanassito/home/pkg/prom"
	"github.com/prometheus/common/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	templateRoot  = flag.String("template-root", "/github/home/pkg/app/templates/", "Root directory of where the templates are stored.")
	liveTemplates = flag.Bool("live-templates", false, "Whether to reload templates on every request")
	airSvcAddr    = flag.String("air-svc", "192.168.1.1:7006", "Address of the Air service.")
	promTimeRx    = regexp.MustCompile("^[0-9]{1,2}[hd]$")
)

func init() {
	flag.Parse()
}

type Slider struct {
	Enabled bool
	Min     int
	Low     int
	High    int
	Max     int
}

func (s Slider) AsPct(v int) int           { return 100 * (v - s.Min) / (s.Max - s.Min) }
func (s Slider) AsPctComplement(v int) int { return 100 - s.AsPct(v) }

func prom2map(promql string, logID string, key string) (data map[string]float64, err error) {
	rs, err := prom.Query(promql, logID)
	if err != nil {
		return data, err
	}
	if rs.Type() != model.ValVector {
		return data, fmt.Errorf("unexpected return type `%s`", rs.Type())
	}
	data = make(map[string]float64, len(rs.(model.Vector)))
	for _, row := range rs.(model.Vector) {
		data[string(row.Metric[model.LabelName(key)])] = float64(row.Value)
	}
	return data, nil
}

type Server struct {
	AirSvc air_proto.AirSvcClient
}

func NewServer() *Server {
	conn, err := grpc.Dial(
		*airSvcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	client := air_proto.NewAirSvcClient(conn)
	return &Server{AirSvc: client}
}

func (s *Server) GetAir() func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	type temperature struct {
		Now float64
		Min float64
		Max float64
	}

	type room struct {
		Name string
		Temp temperature
	}
	var tpl8 *template.Template // Don't use function or force reload on every request
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		since := r.URL.Query().Get("since")
		if !promTimeRx.MatchString(since) {
			since = "1d"
		}
		var err error
		if tpl8 == nil || *liveTemplates {
			root := strings.TrimSuffix(*templateRoot, "/")
			tpl8, err = template.ParseFiles(
				root+"/base.html",
				root+"/air_overview.html",
			)
			if err != nil {
				fmt.Fprintf(w, "Failed to load the templates: %v", err)
				return
			}
		}
		issues := []string{}
		// TODO: execute in parallel
		tempNow, err := prom2map("round(mqtt_temperature{type=\"air\"}, 0.1)", "tempNow", "location")
		if err != nil {
			issues = append(issues, fmt.Sprintf("Failed to load current temperatures: %v", err))
		}
		tempMin, err := prom2map("round(min_over_time(mqtt_temperature{type=\"air\"}[1d]), 0.1)", "tempMin", "location")
		if err != nil {
			issues = append(issues, fmt.Sprintf("Failed to load min temperatures: %v", err))
		}
		tempMax, err := prom2map("round(max_over_time(mqtt_temperature{type=\"air\"}[1d]), 0.1)", "tempMin", "location")
		if err != nil {
			issues = append(issues, fmt.Sprintf("Failed to load max temperatures: %v", err))
		}
		tpl8.ExecuteTemplate(w, "base", struct {
			Issues     []string
			Zaya       room
			Parent     room
			Office     room
			Livingroom room
			Outside    temperature
			Toilettes  temperature
		}{
			Issues: issues,
			Zaya: room{
				Name: "Zaya",
				Temp: temperature{
					Now: tempNow["zaya"],
					Min: tempMin["zaya"],
					Max: tempMax["zaya"],
				},
			},
			Parent: room{
				Name: "Parent",
				Temp: temperature{
					Now: tempNow["parent"],
					Min: tempMin["parent"],
					Max: tempMax["parent"],
				},
			},
			Office: room{
				Name: "Office",
				Temp: temperature{
					Now: tempNow["office"],
					Min: tempMin["office"],
					Max: tempMax["office"],
				},
			},
			Livingroom: room{
				Name: "Livingroom",
				Temp: temperature{
					Now: tempNow["livingroom"],
					Min: tempMin["livingroom"],
					Max: tempMax["livingroom"],
				},
			},
			Outside: temperature{
				Now: tempNow["backyard"],
				Min: tempMin["backyard"],
				Max: tempMax["backyard"],
			},
			Toilettes: temperature{
				Now: tempNow["toilettes"],
				Min: tempMin["toilettes"],
				Max: tempMax["toilettes"],
			},
		})
	}
}

func renderRoom(w http.ResponseWriter, roomID string, serverState *air_proto.ServerState, issues []string) {
	root := strings.TrimSuffix(*templateRoot, "/")
	tpl8, err := template.ParseFiles(
		root+"/base.html",
		root+"/air_room.html",
		root+"/slider.html",
	)
	if err != nil {
		fmt.Fprintf(w, "Failed to load the templates: %v", err)
		return
	}

	if serverState == nil {
		issues = append(issues, "No data from the Air service.")
		serverState = &air_proto.ServerState{
			Outside: &air_proto.Sensor{},
			Rooms:   map[string]*air_proto.Room{},
			Hvacs:   map[string]*air_proto.Hvac{},
		}
	}
	roomState, ok := serverState.Rooms[roomID]
	if !ok {
		issues = append(issues, fmt.Sprintf("No room `%s`", roomID))
		roomState = &air_proto.Room{}
	}
	tpl8.ExecuteTemplate(w, "base", struct {
		Issues     []string
		Name       string
		TempSlider Slider
	}{
		Issues: issues,
		Name:   strings.ToUpper(roomID[0:1]) + roomID[1:],
		TempSlider: Slider{
			Enabled: false,
			Min:     17,
			Low:     int(roomState.GetDesiredTemperatureRange().GetMin()),
			High:    int(roomState.GetDesiredTemperatureRange().GetMax()),
			Max:     35,
		},
	})
}

func (s *Server) GetRoom() func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		issues := []string{}
		resp, err := s.AirSvc.GetState(context.Background(), &air_proto.ReqGetState{})
		if err != nil {
			issues = append(issues, fmt.Sprintf("Error taking to Air: %v", err))
		}
		renderRoom(w, strings.ToLower(params.ByName("roomID")), resp, issues)
	}
}

func (s *Server) PostRoom() func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		issues := []string{}
		roomID := strings.ToLower(params.ByName("roomID"))

		err := r.ParseForm()
		if err != nil {
			issues = append(issues, fmt.Sprintf("Can't parse submitted settings: %v", err))
			resp, err := s.AirSvc.GetState(context.Background(), &air_proto.ReqGetState{})
			if err != nil {
				issues = append(issues, fmt.Sprintf("Error taking to Air: %v", err))
			}
			renderRoom(w, strings.ToLower(params.ByName("roomID")), resp, issues)
		}

		update := &air_proto.ServerState{
			Outside: &air_proto.Sensor{},
			Rooms: map[string]*air_proto.Room{
				roomID: {},
			},
			Hvacs: map[string]*air_proto.Hvac{},
		}

		canSpecifyTempRange := true
		switch strings.ToLower(r.PostForm.Get("isScheduleActive")) {
		case "on":
			update.Rooms[roomID].Schedule = update.Rooms[roomID].GetSchedule()
			t := true
			update.Rooms[roomID].Schedule.IsActive = &t
			canSpecifyTempRange = false
		case "": // checkboxes don't send anything when unchecked :(
			update.Rooms[roomID].Schedule = update.Rooms[roomID].GetSchedule()
			f := false
			update.Rooms[roomID].Schedule.IsActive = &f
		default:
			issues = append(issues, "unknown value for isScheduleActive")
		}

		if canSpecifyTempRange {
			low, errLow := strconv.ParseFloat(r.PostForm.Get("low"), 64)
			high, errHigh := strconv.ParseFloat(r.PostForm.Get("high"), 64)
			if errLow != nil || errHigh != nil {
				issues = append(issues, fmt.Sprintf("Failed to parse `low`(%v) or `high`(%v)", errLow, errHigh))
			} else {
				update.Rooms[roomID].DesiredTemperatureRange = &air_proto.TemperatureRange{
					Min: low,
					Max: high,
				}
			}
		}

		resp, err := s.AirSvc.SetState(context.Background(), update)
		if err != nil {
			issues = append(issues, fmt.Sprintf("Error taking to Air: %v", err))
		}
		renderRoom(w, strings.ToLower(params.ByName("roomID")), resp, issues)
	}
}
