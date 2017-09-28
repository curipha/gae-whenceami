package main

import (
  "fmt"
  "html/template"
  "net/http"
  "strings"
  "time"
)

const textplain = "text/plain"

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
}

func init() {
  http.HandleFunc("/", top)

  http.HandleFunc("/ip", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", textplain)
    fmt.Fprintln(w, ip(r))
  })
  http.HandleFunc("/ua", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", textplain)
    fmt.Fprintln(w, ua(r))
  })
  http.HandleFunc("/country", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", textplain)
    fmt.Fprintln(w, country(r))
  })
  http.HandleFunc("/region", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", textplain)
    fmt.Fprintln(w, region(r))
  })
  http.HandleFunc("/city", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", textplain)
    fmt.Fprintln(w, city(r))
  })
  http.HandleFunc("/time", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", textplain)
    fmt.Fprintln(w, utime(time.Now()))
  })
  http.HandleFunc("/now", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", textplain)
    fmt.Fprintln(w, now(r, time.Now()))
  })
}

func ip(r *http.Request) string {
  return r.RemoteAddr
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
  if r.URL.Path != "/" {
    http.Error(w, "Not Found", http.StatusNotFound)
    return
  }

  t := time.Now()
  template.Must(template.ParseFiles("template.html")).Execute(w, &Parameter {
    Host:      r.Host,
    JST:       timef(t, time.FixedZone("JST", 9*60*60)),
    IP:        ip(r),
    UserAgent: ua(r),
    Country:   country(r),
    Region:    region(r),
    City:      city(r),
    UnixTime:  utime(t),
    Now:       now(r, t),
  })
}
