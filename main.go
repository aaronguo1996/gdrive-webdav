package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/mikea/gdrive-webdav/gdrive"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/net/webdav"
)

var (
	addr         = flag.String("addr", ":8765", "WebDAV service address")
	clientID     = flag.String("client-id", "", "OAuth client id")
	clientSecret = flag.String("client-secret", "", "OAuth client secret")
	debug        = flag.Bool("debug", false, "print debug info")
	trace        = flag.Bool("trace", false, "print trace info")
)

func main() {
	flag.Parse()

	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	if *debug {
		log.SetLevel(log.DebugLevel)
	}
	if *trace {
		log.SetLevel(log.TraceLevel)
	}

	if *clientID == "" {
		fmt.Fprintln(os.Stderr, "--client-id is not specified. See https://developers.google.com/drive/quickstart-go for step-by-step guide.")
		os.Exit(-1)
	}

	if *clientSecret == "" {
		fmt.Fprintln(os.Stderr, "--client-secret is not specified. See https://developers.google.com/drive/quickstart-go for step-by-step guide.")
		os.Exit(-1)
	}

	handler := &webdav.Handler{
		FileSystem: gdrive.NewFS(context.Background(), *clientID, *clientSecret),
		LockSystem: gdrive.NewLS(),
		Logger: func(req *http.Request, err error) {
			log.Tracef("%+v", req)
			if log.IsLevelEnabled(log.TraceLevel) && err != nil {
				log.Errorf("response error %v", err)
			}
		},
	}

	http.HandleFunc("/debug/gc", gcHandler)
	http.HandleFunc("/favicon.ico", notFoundHandler)
	http.HandleFunc("/", handler.ServeHTTP)

	log.Info("Listening on: ", *addr)

	err := http.ListenAndServeTLS(*addr, "/etc/letsencrypt/live/eevee.ucsd.edu/fullchain.pem", "/etc/letsencrypt/live/eevee.ucsd.edu/privkey.pem", nil)
	if err != nil {
		log.Errorf("Error starting HTTP server: %v", err)
		os.Exit(-1)
	}
}

func gcHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("GC")
	runtime.GC()
	w.WriteHeader(200)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
}
