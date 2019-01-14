package main

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type ServerOptions struct {
	Addr string
}

type Server struct {
	store   *Store
	parser  *Parser
	dicts   *Dicts
	options *ServerOptions
	stats   ServerStats
}

func NewServer(store *Store, parser *Parser, dicts *Dicts, options *ServerOptions) *Server {
	return &Server{
		store:   store,
		parser:  parser,
		dicts:   dicts,
		options: options,
		stats: ServerStats{
			Routes: make(ServerStatsRoutes),
		},
	}
}

func (server *Server) GetStats() *ServerStats {
	return &server.stats
}

func (server *Server) Handle() error {
	handleRouteGet("/accounts/filter/", func(writer http.ResponseWriter, request *http.Request) {
		startTime := time.Now()

		filter := NewFilter(server.parser, server.dicts)
		err := filter.Parse(request.URL.RawQuery)
		if err != nil {
			// fmt.Println("Error filter parsing query in", filter.QueryID)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		accounts := server.store.FilterAll(filter)
		encoded := server.parser.EncodeAccounts(accounts, filter)

		writer.Header().Set("Content-Length", strconv.Itoa(len(encoded)))
		writer.WriteHeader(http.StatusOK)
		writer.Write(encoded)

		server.stats.Add(ServerStatsGetFilter, request, time.Now().Sub(startTime))
	})

	// handleRouteGet("/accounts/group/", func(writer http.ResponseWriter, request *http.Request) {
	// 	startTime := time.Now()

	// 	group := NewGroup(server.parser, server.dicts)
	// 	err := group.Parse(request.URL.RawQuery)
	// 	if err != nil {
	// 		// fmt.Println("Error group parsing query in", filter.QueryID)
	// 		writer.WriteHeader(http.StatusBadRequest)
	// 		return
	// 	}

	// 	// writer.WriteHeader(http.StatusOK)
	// 	// writer.Write([]byte(`{"accounts":[]}`))

	// 	aggregation := server.store.GroupAll(group)
	// 	encoded := server.parser.EncodeAggregation(aggregation)

	// 	writer.Header().Set("Content-Length", strconv.Itoa(len(encoded)))
	// 	writer.WriteHeader(http.StatusOK)
	// 	writer.Write(encoded)

	// 	server.stats.Add(ServerStatsGetGroup, request, time.Now().Sub(startTime))
	// })

	handleRoutePost("/accounts/new/", func(writer http.ResponseWriter, request *http.Request) {
		startTime := time.Now()

		rawAccount, err := server.parser.DecodeAccount(request.Body, false)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		_, err = server.store.Add(rawAccount, true, true)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		writer.WriteHeader(http.StatusCreated)
		writer.Write([]byte(`{}`))

		server.stats.Add(ServerStatsPostNew, request, time.Now().Sub(startTime))
	})

	handleRoutePost("/accounts/likes/", func(writer http.ResponseWriter, request *http.Request) {
		startTime := time.Now()

		rawLikes, err := server.parser.DecodeLikes(request.Body)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		err = server.store.AddLikes(rawLikes, true)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		writer.WriteHeader(http.StatusAccepted)
		writer.Write([]byte(`{}`))

		server.stats.Add(ServerStatsPostLikes, request, time.Now().Sub(startTime))
	})

	// recommendRegexp := regexp.MustCompile("^/accounts/([0-9]+)/recommend/$")
	// suggestRegexp := regexp.MustCompile("^/accounts/([0-9]+)/suggest/$")
	updateRegex := regexp.MustCompile("^/accounts/([0-9]+)/$")

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		// recommendMatches := recommendRegexp.FindStringSubmatch(request.URL.Path)
		// suggestMatches := suggestRegexp.FindStringSubmatch(request.URL.Path)
		updateMatches := updateRegex.FindStringSubmatch(request.URL.Path)

		switch {
		// case len(recommendMatches) > 0:
		// 	if request.Method != http.MethodGet {
		// 		writer.WriteHeader(http.StatusMethodNotAllowed)
		// 		return
		// 	}

		// 	// server.stats.Add(request)

		// 	recommend := NewRecommend(server.store, server.dicts)
		// 	err := recommend.Parse(recommendMatches[1], request.URL.RawQuery)
		// 	if err != nil {
		// 		writer.WriteHeader(http.StatusBadRequest)
		// 		return
		// 	}
		// 	if recommend.Account == nil {
		// 		writer.WriteHeader(http.StatusNotFound)
		// 		return
		// 	}

		// 	writer.WriteHeader(http.StatusOK)
		// 	writer.Write([]byte(`{"accounts":[]}`))
		// 	return
		// case len(suggestMatches) > 0:
		// 	if request.Method != http.MethodGet {
		// 		writer.WriteHeader(http.StatusMethodNotAllowed)
		// 		return
		// 	}

		// 	// server.stats.Add(request)

		// 	suggest := NewSuggest(server.store, server.dicts)
		// 	err := suggest.Parse(suggestMatches[1], request.URL.RawQuery)
		// 	if err != nil {
		// 		writer.WriteHeader(http.StatusBadRequest)
		// 		return
		// 	}
		// 	if suggest.Account == nil {
		// 		writer.WriteHeader(http.StatusNotFound)
		// 		return
		// 	}

		// 	writer.WriteHeader(http.StatusOK)
		// 	writer.Write([]byte(`{"accounts":[]}`))
		// 	return
		case len(updateMatches) > 0:
			startTime := time.Now()

			if request.Method != http.MethodPost {
				writer.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			ui64, err := strconv.ParseUint(updateMatches[1], 10, 32)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				return
			}

			account, ok := server.store.Get(ID(ui64))
			if !ok {
				writer.WriteHeader(http.StatusNotFound)
				return
			}

			rawAccount, err := server.parser.DecodeAccount(request.Body, true)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				return
			}

			_, err = server.store.Update(account.ID, rawAccount, true)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				return
			}

			writer.WriteHeader(http.StatusAccepted)
			writer.Write([]byte(`{}`))

			server.stats.Add(ServerStatsPostUpdate, request, time.Now().Sub(startTime))

			return
		}

		writer.WriteHeader(http.StatusNotFound)
	})

	return http.ListenAndServe(server.options.Addr, nil)
}

type ServerStats struct {
	Total  uint64
	Routes ServerStatsRoutes
}

type ServerStatsRoutes map[string]ServerStatsRoute

type ServerStatsRoute []ServerStatsSet

type ServerStatsSet struct {
	Params string
	Total  uint64
	Time   time.Duration
}

const (
	ServerStatsGetFilter  = "/accounts/filter/"
	ServerStatsGetGroup   = "/accounts/group/"
	ServerStatsPostNew    = "/accounts/new/"
	ServerStatsPostUpdate = "/accounts/XXX/"
	ServerStatsPostLikes  = "/accounts/likes/"
)

func (stats *ServerStats) Add(path string, request *http.Request, d time.Duration) {
	var p sort.StringSlice
	for param := range request.URL.Query() {
		if param == "query_id" || param == "limit" {
			continue
		}
		p = append(p, param)
	}
	params := "<empty>"
	if request.Method == http.MethodGet {
		if len(p) > 0 {
			sort.Sort(p)
			params = strings.Join(p, ",")
		}
	}
	if _, ok := stats.Routes[path]; !ok {
		stats.Routes[path] = make(ServerStatsRoute, 0, 512)
	}
	index := len(stats.Routes[path])
	for indx, set := range stats.Routes[path] {
		if set.Params == params {
			index = indx
		}
	}
	if index != len(stats.Routes[path]) {
		stats.Routes[path][index].Total++
		stats.Routes[path][index].Time += d
	} else {
		stats.Routes[path] = append(stats.Routes[path], ServerStatsSet{
			Params: params,
			Total:  1,
			Time:   d,
		})
	}
	stats.Total++
}

// func (stats *ServerStats) Routes() ServerStatsRoutes {
// 	return stats.routes
// }

func (stats *ServerStats) Sort() {
	for routeName := range stats.Routes {
		sort.Sort(stats.Routes[routeName])
	}
}

func (stats *ServerStats) Format() string {
	str := fmt.Sprintf("TotalRequests = %d", stats.Total)
	for routeName := range stats.Routes {
		str += fmt.Sprintf("\n%s:", routeName)
		for _, set := range stats.Routes[routeName] {
			str += fmt.Sprintf("\n    %s, total = %d, total_time = %d, avg = %dms", set.Params, set.Total, set.Time, set.Time/time.Millisecond/time.Duration(set.Total))
		}
	}
	return str
}

func (statsRoute ServerStatsRoute) Len() int { return len(statsRoute) }
func (statsRoute ServerStatsRoute) Swap(i, j int) {
	statsRoute[i], statsRoute[j] = statsRoute[j], statsRoute[i]
}
func (statsRoute ServerStatsRoute) Less(i, j int) bool {
	return uint64(statsRoute[i].Time)/statsRoute[i].Total > uint64(statsRoute[j].Time)/statsRoute[j].Total
}

func handleRoute(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		if len(pattern) != len(request.URL.Path) {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		handler(writer, request)
	})
}

func handleRouteGet(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	handleRoute(pattern, func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler(writer, request)
	})
}

func handleRoutePost(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	handleRoute(pattern, func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler(writer, request)
	})
}
