package main

import (
  "errors"
  "flag"
  "fmt"
  "html/template"
  "log"
  "net"
  "net/http"
  "net/url"
  "strings"

  "tailscale.com/client/tailscale"
  "tailscale.com/hostinfo"
  "tailscale.com/ipn"
  "tailscale.com/tsnet"
)

var (
  verbose    = flag.Bool("verbose", false, "be verbose")
  controlURL = flag.String("control-url", ipn.DefaultControlURL, "the URL base of the control plane (i.e. coordination server)")
  target     = flag.String("target", "", "the url to redirect to")
  dev        = flag.String("dev-listen", "", "if non-empty, listen on this addr and don't use tsnet")
  hostname   = flag.String("hostname", "", "service name")
)

var localClient *tailscale.LocalClient

func main() {
  if err := Run(); err != nil {
    log.Fatal(err)
  }
}

func Run() error {
  flag.Parse()

  hostinfo.SetApp("redirect")

  if *target == "" {
    return errors.New("--target cannot be empty")
  }

  targetUrl, err := url.Parse(*target)
  if err != nil {
    return fmt.Errorf("Unable to parse target url: %w", err)
  }

  http.HandleFunc("/", serveRedirect(*targetUrl))

  if *dev != "" {
    // override default hostname for dev mode
    if *hostname == "" {
      if h, p, err := net.SplitHostPort(*dev); err == nil {
        if h == "" {
          h = "localhost"
        }
        *hostname = fmt.Sprintf("%s:%s", h, p)
        println("hostname set")
      } else {
        log.Fatal("unable to listen on %s", *dev)
      }
    }

    log.Printf("Running in dev mode on %s ...", *dev)
    log.Fatal(http.ListenAndServe(*dev, nil))
  }

  if *hostname == "" {
    return errors.New("--hostname cannot be empty")
  }

  srv := &tsnet.Server{
    Dir:        ".tsnet-state",
    ControlURL: *controlURL,
    Hostname:   *hostname,
    Logf: func(format string, args ...any) {
      // Show the log line with the interactive tailscale login link even when verbose is off
      if strings.Contains(format, "To start this tsnet server") {
        log.Printf(format, args...)
      }
    },
  }
  if *verbose {
    srv.Logf = log.Printf
  }
  if err := srv.Start(); err != nil {
    return err
  }
  localClient, _ = srv.LocalClient()

  l80, err := srv.Listen("tcp", ":80")
  if err != nil {
    return err
  }

  log.Printf("Serving http://%s/ ...", *hostname)
  if err := http.Serve(l80, nil); err != nil {
    return err
  }
  return nil
}

var (
  homeTmpl    *template.Template
  receiveTmpl *template.Template
)

func serveRedirect(targetUrl url.URL) func(w http.ResponseWriter, r *http.Request) {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" {
      http.Error(w, "HTTP Method Unsupported", http.StatusBadRequest)
      return
    }
    newUrl := r.URL
    newUrl.Scheme = targetUrl.Scheme
    newUrl.Host = targetUrl.Host
    newUrl.Path = targetUrl.Path + r.URL.Path[1:len(r.URL.Path)]
    http.Redirect(w, r, newUrl.String(), http.StatusFound)
  }
}

func devMode() bool { return *dev != "" }
