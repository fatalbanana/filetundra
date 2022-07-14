package web

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/fatalbanana/filetundra/internal/env"

	"github.com/gorilla/mux"
)

var (
	Server *http.Server
)

func RunWebserver() error {
	router := mux.NewRouter()
	router.PathPrefix("/browse").HandlerFunc(browseHandler)
	router.PathPrefix("/download").HandlerFunc(downloadHandler)
	router.HandleFunc("/search", searchHandler)
	router.PathPrefix("/static").HandlerFunc(staticHandler)
	Server = &http.Server{
		Addr:              net.JoinHostPort(env.Env.HTTPAddress, fmt.Sprintf("%d", env.Env.HTTPPort)),
		Handler:           router,
		ReadHeaderTimeout: 1 * time.Second,
		ReadTimeout:       5 * time.Second,
	}
	return Server.ListenAndServe()
}
