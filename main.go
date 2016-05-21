package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Luzifer/go-metar"
	"github.com/Luzifer/go_helpers/str"
	"github.com/Luzifer/rconfig"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/robfig/cron"
)

var (
	cfg = struct {
		Listen              string   `flag:"listen" default:":3000" description:"IP/Port to listen on"`
		Stations            []string `flag:"station,s" description:"Stations (airport codes) to query"`
		QueryInterval       string   `flag:"interval" default:"5m" description:"Interval to fetch the data"`
		parsedQueryInterval time.Duration
	}{}

	version = "dev"

	labelNames = []string{"station"}
	metrics    = map[string]*prometheus.GaugeVec{}
)

const (
	metricTemperature   = "temperature"
	metricSuccess       = "query_success"
	metricFetchTime     = "fetch_time"
	metricTime          = "observation_time"
	metricDewpoint      = "dewpoint"
	metricWindDirection = "wind_direction"
	metricWindSpeed     = "wind_speed"
	metricVisibility    = "visibility"
	metricAltimeter     = "altimeter"
	metricSkyCover      = "skycover"
	metricBft           = "wind_force"
)

func buildMetric(name, help string) {
	metrics[name] = prometheus.MustRegisterOrGet(prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "metar",
		Name:      name,
		Help:      help,
	}, labelNames)).(*prometheus.GaugeVec)
}

func registerMetrics() {
	buildMetric(metricTemperature, "Air temperature (celsius)")
	buildMetric(metricSuccess, "Indicates whether the last fetch was a success (0/1)")
	buildMetric(metricTime, "Contains the observation time of the current data reported by the station (UTC)")
	buildMetric(metricDewpoint, "Dewpoint temperature (celsius)")
	buildMetric(metricWindDirection, "Direction from which the wind is blowing. 0 degrees=variable wind direction.")
	buildMetric(metricWindSpeed, "Wind speed; 0 degree wdir and 0 wspd = calm winds (km/h)")
	buildMetric(metricVisibility, "Horizontal visibility (km)")
	buildMetric(metricAltimeter, "Altimeter (hPa)")
	buildMetric(metricSkyCover, "Sky cover in % (0 = clear, 1 = full cover)")
	buildMetric(metricFetchTime, "Contains the timestamp of the last successful fetch")
	buildMetric(metricBft, "Wind force in Beaufort wind force scale")
}

func main() {
	registerMetrics()

	var err error
	if err = rconfig.Parse(&cfg); err != nil {
		log.Fatalf("Error while parsing flags: %s", err)
	}

	if cfg.parsedQueryInterval, err = time.ParseDuration(cfg.QueryInterval); err != nil {
		log.Fatalf("Unable to parse interval parameter: %s", err)
	}

	if len(cfg.Stations) == 1 && cfg.Stations[0] == "" {
		log.Fatalf("You need to specify at least one station to fetch data from.")
	}

	c := cron.New()
	c.AddFunc(fmt.Sprintf("@every %s", cfg.parsedQueryInterval), queryMetrics)
	c.Start()

	queryMetrics()

	r := mux.NewRouter()
	r.Handle("/metrics", prometheus.Handler())
	r.HandleFunc("/", func(res http.ResponseWriter, r *http.Request) {
		http.Error(res, "I'm fine but for metrics visit /metrics", http.StatusOK)
	})

	log.Fatalf("[ERR] %s", http.ListenAndServe(cfg.Listen, r))
}

func queryMetrics() {
	for i := range cfg.Stations {
		go queryStation(cfg.Stations[i])
	}
}

func queryStation(station string) {
	data, err := metar.FetchCurrentStationWeather(station)
	if err != nil {
		metrics[metricSuccess].WithLabelValues(station).Set(0)
		log.Printf("[ERR] Unable to fetch data for station %s: %s", station, err)
		return
	}

	metrics[metricTemperature].WithLabelValues(station).Set(data.Temperature)
	metrics[metricSuccess].WithLabelValues(station).Set(1)
	metrics[metricTime].WithLabelValues(station).Set(float64(data.ObservationTime.UTC().Unix()))
	metrics[metricDewpoint].WithLabelValues(station).Set(data.Dewpoint)
	metrics[metricWindDirection].WithLabelValues(station).Set(float64(data.WindDirDegrees))
	metrics[metricWindSpeed].WithLabelValues(station).Set(metar.KtsToMs(float64(data.WindSpeed)) * 3.6)
	metrics[metricVisibility].WithLabelValues(station).Set(metar.StatMileToKm(data.VisibilityStatute))
	metrics[metricAltimeter].WithLabelValues(station).Set(metar.InHgTohPa(data.Altimeter))

	switch {
	case str.StringInSlice(string(data.SkyCondition.SkyCover), []string{"SKC", "CLR", "NSC", "CAVOK"}):
		metrics[metricSkyCover].WithLabelValues(station).Set(0)
	case data.SkyCondition.SkyCover == metar.SkyCoverFEW:
		metrics[metricSkyCover].WithLabelValues(station).Set(2.0 / 8.0)
	case data.SkyCondition.SkyCover == metar.SkyCoverSCT:
		metrics[metricSkyCover].WithLabelValues(station).Set(4.0 / 8.0)
	case data.SkyCondition.SkyCover == metar.SkyCoverBKN:
		metrics[metricSkyCover].WithLabelValues(station).Set(7.0 / 8.0)
	case data.SkyCondition.SkyCover == metar.SkyCoverOVC:
		metrics[metricSkyCover].WithLabelValues(station).Set(1)
	}

	metrics[metricFetchTime].WithLabelValues(station).Set(float64(time.Now().UTC().Unix()))
	metrics[metricBft].WithLabelValues(station).Set(float64(metar.KtsToBft(float64(data.WindSpeed))))
}
