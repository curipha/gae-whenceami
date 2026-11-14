package main

import (
  "fmt"
  "html/template"
  "log"
  "net/http"
  "os"
  "strconv"
  "strings"
  "time"

  _ "time/tzdata"
)

var tmpl = template.Must(template.ParseFiles("template.html"))

type Parameter struct {
  Host      string
  IP        string
  UserAgent string
  Country   string
  Region    string
  City      string
  UnixTime  int64
  Now       string
  JST       string
  Latitude  string
  Longitude string
  MapNorth  float64
  MapSouth  float64
  MapEast   float64
  MapWest   float64
}


func main() {
  mux := http.NewServeMux()

  // "GET" also accepts HEAD; other methods return 405 and an Allow header.
  mux.HandleFunc("GET /{$}", top)

  mux.HandleFunc("GET /ip", plain(func(r *http.Request) any { return ip(r) }))
  mux.HandleFunc("GET /ua", plain(func(r *http.Request) any { return ua(r) }))
  mux.HandleFunc("GET /country", plain(func(r *http.Request) any { return country(r) }))
  mux.HandleFunc("GET /region", plain(func(r *http.Request) any { return region(r) }))
  mux.HandleFunc("GET /city", plain(func(r *http.Request) any { return city(r) }))
  mux.HandleFunc("GET /time", plain(func(r *http.Request) any { return utime(time.Now()) }))
  mux.HandleFunc("GET /now", plain(func(r *http.Request) any { return now(r, time.Now()) }))

  port := os.Getenv("PORT")
  if port == "" {
    port = "8080"
    log.Printf("Defaulting to port %s", port)
  }

  server := &http.Server{
    Addr:              ":" + port,
    Handler:           mux,
    ReadHeaderTimeout: 10 * time.Second, // Mitigate Slowloris
    ReadTimeout:       10 * time.Second,
    WriteTimeout:      10 * time.Second,
    IdleTimeout:       60 * time.Second,
  }

  log.Printf("Listening on port %s", port)
  err := server.ListenAndServe()
  if err != nil {
    log.Fatal(err)
  }
}

func plain(f func(*http.Request) any) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain")
    fmt.Fprintln(w, f(r))
  }
}

func ip(r *http.Request) string {
  addr := r.Header.Get("X-Appengine-User-Ip")
  if addr == "" {
    log.Println("'X-Appengine-User-Ip' header is empty! Get remote address from 'X-Forwarded-For'.")
    addr = strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]
  }
  return strings.TrimSpace(addr)
}
func ua(r *http.Request) string {
  return strings.TrimSpace(r.Header.Get("User-Agent"))
}
func country(r *http.Request) string {
  return strings.TrimSpace(r.Header.Get("X-AppEngine-Country"))
}
func region(r *http.Request) string {
  return strings.TrimSpace(r.Header.Get("X-AppEngine-Region"))
}
func city(r *http.Request) string {
  return strings.TrimSpace(r.Header.Get("X-AppEngine-City"))
}
func utime(t time.Time) int64 {
  return t.Unix()
}
func timef(t time.Time, zone *time.Location) string {
  return t.In(zone).Format(time.RFC1123)
}
func now(r *http.Request, t time.Time) string {
  zone, err := time.LoadLocation(strings.TrimSpace(r.URL.Query().Get("zone")))
  if err != nil {
    zone = time.UTC
  }
  return timef(t, zone)
}

func top(w http.ResponseWriter, r *http.Request) {
  agent := strings.ToLower(strings.SplitN(ua(r), "/", 2)[0])
  if agent == "curl" || agent == "wget" {
    w.Header().Set("Content-Type", "text/plain")
    fmt.Fprintln(w, ip(r))
  } else {
    t := time.Now()

    var north, south, east, west float64
    latitude, longitude, found := strings.Cut(strings.TrimSpace(r.Header.Get("X-Appengine-CityLatLong")), ",")
    if found {
      maplat, errlat := strconv.ParseFloat(latitude, 64)
      maplon, errlon := strconv.ParseFloat(longitude, 64)
      if errlat == nil && errlon == nil {
        north = maplat + 0.05
        south = maplat - 0.05
        east = maplon + 0.05
        west = maplon - 0.05
      } else {
        // Reset lat/long if it failed to convert to float64
        latitude = "0.000000"
        longitude = "0.000000"
      }
    } else {
      // GAE sets lat/long to "0.000000" when the city cannot be determined
      // https://docs.cloud.google.com/appengine/docs/standard/reference/request-headers?tab=go#app_engine-specific_headers
      latitude = "0.000000"
      longitude = "0.000000"
    }

    err := tmpl.Execute(w, &Parameter {
      Host:      r.Host,
      JST:       timef(t, time.FixedZone("JST", 9*60*60)),
      IP:        ip(r),
      UserAgent: ua(r),
      Country:   country(r),
      Region:    region(r),
      City:      city(r),
      UnixTime:  utime(t),
      Now:       now(r, t),
      Latitude:  latitude,
      Longitude: longitude,
      MapNorth:  north,
      MapSouth:  south,
      MapEast:   east,
      MapWest:   west,
    })
    if err != nil {
      log.Println(err)
    }
  }
}
