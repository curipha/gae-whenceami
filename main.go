package main

import (
  "fmt"
  "html/template"
  "log"
  "net/http"
  "os"
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

func main() {
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

  port := os.Getenv("PORT")
  if port == "" {
    port = "8080"
    log.Printf("Defaulting to port %s", port)
  }

  log.Printf("Listening on port %s", port)
  err := http.ListenAndServe(":" + port, nil)
  if err != nil {
    log.Fatal(err)
  }
}

func ip(r *http.Request) string {
  addr := r.Header.Get("X-Appengine-User-Ip")
  if addr == "" {
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
  if r.URL.Path != "/" {
    http.NotFound(w, r)
    return
  }

  agent := strings.ToLower(strings.SplitN(ua(r), "/", 2)[0])

  if agent == "curl" || agent == "wget" {
    w.Header().Set("Content-Type", textplain)
    fmt.Fprintln(w, ip(r))
  } else {
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
}
