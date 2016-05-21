[![Download on GoBuilder](http://badge.luzifer.io/v1/badge?title=Download%20on&text=GoBuilder)](https://gobuilder.me/github.com/Luzifer/prometheus-metar)
[![License: Apache v2.0](https://badge.luzifer.io/v1/badge?color=5d79b5&title=license&text=Apache+v2.0)](http://www.apache.org/licenses/LICENSE-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/Luzifer/prometheus-metar)](https://goreportcard.com/report/github.com/Luzifer/prometheus-metar)

# Luzifer / prometheus-metar

`prometheus-metar` is a data exporter for [prometheus](https://prometheus.io/) to fetch weather condition data from METAR stations. Those information are usually used for aviation purposes and are refreshed about every 30m.

## Usage

### Standalone binary

1. Fetch the right binary from [GoBuilder](https://gobuilder.me/github.com/Luzifer/prometheus-metar) or build it from source
2. Search for the [nearest airport code](https://www.world-airport-codes.com/) (ICAO code) for the desired location
3. Start the exporter  
```
$ prometheus-metar -s EDDW -s EDDH
```

### Docker image

1. Search for the [nearest airport code](https://www.world-airport-codes.com/) (ICAO code) for the desired location
2. Start the exporter  
```
$ docker run -d quay.io/luzifer/prometheus-metar -s EDDW -s EDDH
```
