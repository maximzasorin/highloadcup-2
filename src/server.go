package main

import (
	"bytes"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

type ServerRoute uint16

var defaultPostResponse = []byte("{}")

const (
	ServerRouteGetFilter ServerRoute = 1 << iota
	ServerRouteGetGroup
	ServerRouteGetRecommend
	ServerRouteGetSuggest
	ServerRoutePostNew
	ServerRoutePostUpdate
	ServerRoutePostLikes
	ServerRouteAllGet  = ServerRouteGetFilter | ServerRouteGetGroup | ServerRouteGetRecommend | ServerRouteGetSuggest
	ServerRouteAllPost = ServerRoutePostNew | ServerRoutePostUpdate | ServerRoutePostLikes
	ServerRouteAll     = ServerRouteAllGet | ServerRouteAllPost
)

// ----

type AccountsBuffer []*Account

var accountsBufferPool = sync.Pool{
	New: func() interface{} {
		ab := make(AccountsBuffer, 0, 50)
		return &ab
	},
}

func BorrowAccountsBuffer() *AccountsBuffer {
	ab := accountsBufferPool.Get().(*AccountsBuffer)
	ab.Reset()
	return ab
}

func (buffer *AccountsBuffer) Reset() {
	*buffer = (*buffer)[:0]
}

func (buffer *AccountsBuffer) Release() {
	accountsBufferPool.Put(buffer)
}

// ----

type GroupsBuffer struct {
	orderAsc bool
	groups   []*GroupEntry
}

var groupsBufferPool = sync.Pool{
	New: func() interface{} {
		return &GroupsBuffer{
			groups: make([]*GroupEntry, 0, 50),
		}
	},
}

func BorrowGroupsBuffer() *GroupsBuffer {
	gb := groupsBufferPool.Get().(*GroupsBuffer)
	gb.Reset()
	return gb
}

func (buffer *GroupsBuffer) Reset() {
	buffer.orderAsc = false
	buffer.groups = buffer.groups[:0]
}

func (buffer *GroupsBuffer) Release() {
	groupsBufferPool.Put(buffer)
}

// ----

type OutputBuffer struct {
	bytes.Buffer
	buf []byte
}

var outputPool = sync.Pool{
	New: func() interface{} {
		return &OutputBuffer{buf: make([]byte, 0, 16*1024)}
	},
}

func (buffer *OutputBuffer) Release() {
	outputPool.Put(buffer)
}

func BorrowBuffer() *OutputBuffer {
	b := outputPool.Get().(*OutputBuffer)
	b.Reset()
	return b
}

type ServerOptions struct {
	Addr   string
	Stats  bool
	Routes ServerRoute
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
	updateRegex := regexp.MustCompile("^/accounts/([0-9]+)/$")
	recommendRegexp := regexp.MustCompile("^/accounts/([0-9]+)/recommend/$")
	suggestRegexp := regexp.MustCompile("^/accounts/([0-9]+)/suggest/$")

	handler := func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())

		// if strings.HasPrefix(path, "/debug") {
		// 	pprofhandler.PprofHandler(ctx)
		// 	return
		// }

		switch path {
		case "/accounts/filter/":
			server.handleFilterRequest(ctx)
		case "/accounts/group/":
			server.handleGroupRequest(ctx)
		case "/accounts/new/":
			server.handleNewRequest(ctx)
		case "/accounts/likes/":
			server.handleLikesRequest(ctx)
		default:
			updateMatches := updateRegex.FindStringSubmatch(path)
			if len(updateMatches) > 0 {
				server.handleUpdateRequest(ctx, updateMatches)
			} else {
				recommendMatches := recommendRegexp.FindStringSubmatch(path)
				if len(recommendMatches) > 0 {
					server.handleRecommendRequest(ctx, recommendMatches)
				} else {
					suggestMatches := suggestRegexp.FindStringSubmatch(path)
					if len(suggestMatches) > 0 {
						server.handleSuggestRequest(ctx, suggestMatches)
					} else {
						ctx.NotFound()
					}
				}
			}
		}
	}
	return fasthttp.ListenAndServe(server.options.Addr, handler)
}

func (srv *Server) handleFilterRequest(ctx *fasthttp.RequestCtx) {
	filter := BorrowFilter(srv.parser, srv.dicts)
	defer filter.Release()

	err := filter.Parse(string(ctx.URI().QueryString()))
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	accounts := BorrowAccountsBuffer()
	defer accounts.Release()

	srv.store.Filter(filter, accounts)

	buffer := BorrowBuffer()
	defer buffer.Release()

	srv.parser.EncodeAccounts(*accounts, buffer, filter)

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyStream(buffer, buffer.Len())

	// if server.options.Stats {
	// 	params := strings.Join(getParams(request, []string{"limit", "query_id"}), ",")
	// 	server.stats.Add(ServerStatsGetFilter, params, st)

	// 	d := time.Now().Sub(st) / time.Microsecond
	// 	if d > 1000 {
	// 		fmt.Printf("long query = %s, %d mus\n", request.URL.RequestURI(), d)
	// 	}
	// }
}

func (srv *Server) handleNewRequest(ctx *fasthttp.RequestCtx) {
	rawAccount := BorrowRawAccount()
	defer rawAccount.Release()

	err := srv.parser.DecodeAccount(ctx.PostBody(), rawAccount, false)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	_, err = srv.store.Add(rawAccount, true, true)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusCreated)
	ctx.Write(defaultPostResponse)
}

func (srv *Server) handleLikesRequest(ctx *fasthttp.RequestCtx) {
	likes := BorrowLikes()
	defer likes.Release()

	err := srv.parser.DecodeLikes(ctx.PostBody(), likes)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	likes.Truncate()

	err = srv.store.AddLikes(likes, true)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusAccepted)
	ctx.Write(defaultPostResponse)
}

func (srv *Server) handleUpdateRequest(ctx *fasthttp.RequestCtx, matches []string) {
	ui64, err := strconv.ParseUint(matches[1], 10, 32)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	account := srv.store.Get(ID(ui64))
	if account == nil {
		ctx.NotFound()
		return
	}

	rawAccount := BorrowRawAccount()
	defer rawAccount.Release()

	err = srv.parser.DecodeAccount(ctx.PostBody(), rawAccount, false)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	_, err = srv.store.Update(account.ID, rawAccount, true)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusAccepted)
	ctx.Write(defaultPostResponse)
}

func (srv *Server) handleGroupRequest(ctx *fasthttp.RequestCtx) {
	group := BorrowGroup(srv.parser, srv.dicts)
	defer group.Release()

	err := group.Parse(string(ctx.URI().QueryString()))
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	groups := BorrowGroupsBuffer()
	defer groups.Release()

	srv.store.Group(group, groups)

	buffer := BorrowBuffer()
	defer buffer.Release()

	srv.parser.EncodeGroupEntries(groups, buffer)

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyStream(buffer, buffer.Len())
}

func (srv *Server) handleSuggestRequest(ctx *fasthttp.RequestCtx, matches []string) {
	ui64, err := strconv.ParseUint(matches[1], 10, 32)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	account := srv.store.Get(ID(ui64))
	if account == nil {
		ctx.NotFound()
		return
	}

	suggest := BorrowSuggest(srv.store, srv.dicts)
	defer suggest.Release()

	err = suggest.Parse(string(ctx.URI().QueryString()))
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	accounts := BorrowAccountsBuffer()
	defer accounts.Release()

	srv.store.Suggest(account, suggest, accounts)

	buffer := BorrowBuffer()
	defer buffer.Release()

	srv.parser.EncodeAccounts(*accounts, buffer, suggest)

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyStream(buffer, buffer.Len())
}

func (srv *Server) handleRecommendRequest(ctx *fasthttp.RequestCtx, matches []string) {
	ui64, err := strconv.ParseUint(matches[1], 10, 32)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	account := srv.store.Get(ID(ui64))
	if account == nil {
		ctx.NotFound()
		return
	}

	recommend := BorrowRecommend(srv.store, srv.dicts)
	defer recommend.Release()

	err = recommend.Parse(string(ctx.URI().QueryString()))
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	accounts := BorrowAccountsBuffer()
	defer accounts.Release()

	srv.store.Recommend(account, recommend, accounts)

	buffer := BorrowBuffer()
	defer buffer.Release()

	srv.parser.EncodeAccounts(*accounts, buffer, recommend)

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyStream(buffer, buffer.Len())
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
