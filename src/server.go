package main

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
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
		// startTime := time.Now()

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

		// params := strings.Join(getParams(request, []string{"limit", "query_id"}), ",")
		// server.stats.Add(ServerStatsGetFilter, params, startTime)
	})

	handleRouteGet("/accounts/group/", func(writer http.ResponseWriter, request *http.Request) {
		// startTime := time.Now()

		group := NewGroup(server.parser, server.dicts)
		err := group.Parse(request.URL.RawQuery)
		if err != nil {
			// fmt.Println("Error group parsing query in", filter.QueryID)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		// writer.WriteHeader(http.StatusOK)
		// writer.Write([]byte(`{"accounts":[]}`))

		groupEntries, orderAsc := server.store.GroupAll(group)
		encoded := server.parser.EncodeGroupEntries(groupEntries, orderAsc)

		writer.Header().Set("Content-Length", strconv.Itoa(len(encoded)))
		writer.WriteHeader(http.StatusOK)
		writer.Write(encoded)

		// params := getParams(request, []string{"limit", "query_id", "order", "keys"})
		// keys, ok := request.URL.Query()["keys"]
		// if ok {
		// 	params = append([]string{"keys:[" + keys[0] + "]"}, params...)
		// }
		// paramsStr := strings.Join(params, ",")
		// server.stats.Add(ServerStatsGetGroup, paramsStr, startTime)

		// d := time.Now().Sub(startTime) / time.Microsecond
		// if d > 1000 {
		// 	fmt.Printf("long query = %s, %d mus\n", request.URL.RequestURI(), d)
		// }
	})

	handleRoutePost("/accounts/new/", func(writer http.ResponseWriter, request *http.Request) {
		// startTime := time.Now()

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

		// server.stats.Add(ServerStatsPostNew, "", startTime)
	})

	handleRoutePost("/accounts/likes/", func(writer http.ResponseWriter, request *http.Request) {
		// startTime := time.Now()

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

		// server.stats.Add(ServerStatsPostLikes, "", startTime)
	})

	recommendRegexp := regexp.MustCompile("^/accounts/([0-9]+)/recommend/$")
	suggestRegexp := regexp.MustCompile("^/accounts/([0-9]+)/suggest/$")
	updateRegex := regexp.MustCompile("^/accounts/([0-9]+)/$")

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		recommendMatches := recommendRegexp.FindStringSubmatch(request.URL.Path)
		suggestMatches := suggestRegexp.FindStringSubmatch(request.URL.Path)
		updateMatches := updateRegex.FindStringSubmatch(request.URL.Path)

		switch {
		case len(recommendMatches) > 0:
			// startTime := time.Now()

			if request.Method != http.MethodGet {
				writer.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			ui64, err := strconv.ParseUint(recommendMatches[1], 10, 32)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				return
			}

			account := server.store.Get(ID(ui64))
			if account == nil {
				writer.WriteHeader(http.StatusNotFound)
				return
			}

			recommend := NewRecommend(server.store, server.dicts)
			err = recommend.Parse(request.URL.RawQuery)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				return
			}

			accounts := server.store.RecommendAll(account, recommend)
			encoded := server.parser.EncodeAccounts(accounts, recommend)

			writer.Header().Set("Content-Length", strconv.Itoa(len(encoded)))
			writer.WriteHeader(http.StatusOK)
			writer.Write(encoded)

			// params := strings.Join(getParams(request, []string{"limit", "query_id"}), ",")
			// server.stats.Add(ServerStatsGetRecommend, params, startTime)

			return
		case len(suggestMatches) > 0:
			// startTime := time.Now()

			if request.Method != http.MethodGet {
				writer.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			ui64, err := strconv.ParseUint(suggestMatches[1], 10, 32)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				return
			}

			account := server.store.Get(ID(ui64))
			if account == nil {
				writer.WriteHeader(http.StatusNotFound)
				return
			}

			suggest := NewSuggest(server.store, server.dicts)
			err = suggest.Parse(request.URL.RawQuery)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				return
			}

			accounts := server.store.SuggestAll(account, suggest)
			encoded := server.parser.EncodeAccounts(accounts, suggest)

			writer.Header().Set("Content-Length", strconv.Itoa(len(encoded)))
			writer.WriteHeader(http.StatusOK)
			writer.Write(encoded)

			// params := strings.Join(getParams(request, []string{"limit", "query_id"}), ",")
			// server.stats.Add(ServerStatsGetSuggest, params, startTime)

			return
		case len(updateMatches) > 0:
			// startTime := time.Now()

			if request.Method != http.MethodPost {
				writer.WriteHeader(http.StatusMethodNotAllowed)
				return
			}

			ui64, err := strconv.ParseUint(updateMatches[1], 10, 32)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				return
			}

			account := server.store.Get(ID(ui64))
			if account == nil {
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

			// server.stats.Add(ServerStatsPostUpdate, "", startTime)

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
	ServerStatsGetFilter    = "/accounts/filter/"
	ServerStatsGetGroup     = "/accounts/group/"
	ServerStatsPostNew      = "/accounts/new/"
	ServerStatsPostUpdate   = "/accounts/XXX/"
	ServerStatsGetRecommend = "/accounts/XXX/recommend/"
	ServerStatsGetSuggest   = "/accounts/XXX/suggest/"
	ServerStatsPostLikes    = "/accounts/likes/"
)

func (stats *ServerStats) Add(path string, params string, start time.Time) {
	d := time.Now().Sub(start)
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

func getParams(request *http.Request, except []string) []string {
	var params sort.StringSlice
	for param := range request.URL.Query() {
		skip := false
		for _, ex := range except {
			if ex == param {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		params = append(params, param)
	}
	if len(params) > 0 {
		sort.Sort(params)
	}
	return params
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
			str += fmt.Sprintf("\n    %s, total = %d, total_time = %d, avg = %dmus", set.Params, set.Total, set.Time, set.Time/time.Microsecond/time.Duration(set.Total))
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
