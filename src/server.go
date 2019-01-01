package main

import (
	"net/http"
)

type ServerOptions struct {
	Addr string
}

type Server struct {
	store   *Store
	options *ServerOptions
}

func NewServer(store *Store, options *ServerOptions) *Server {
	return &Server{store, options}
}

func (server *Server) Handle() error {
	http.HandleFunc("/accounts/filter/", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		filter := NewFilter()
		err := filter.Parse(request.URL.RawQuery)
		if err != nil {
			// fmt.Println("Error parsing query in", filter.QueryID)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		server.store.FilterAll(filter, writer)
	})

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNotFound)
	})

	return http.ListenAndServe(":80", nil)
}
