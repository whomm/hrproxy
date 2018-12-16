package httptool

import (
	"errors"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
)

func NewServer(add string, backservice string) *LServeMux {
	rpx, err := ReverseProxy(backservice)
	if err != nil {
		panic(err)
	}
	return &LServeMux{rp: rpx, add: add}
}

// LimitHandler is a middleware that performs rate-limiting given http.Handler struct.
func LimitHandler(path string, lmt *limiter.Limiter, next http.Handler) http.Handler {
	middle := func(w http.ResponseWriter, r *http.Request) {
		if lmt.LimitReached(path) {
			lmt.ExecOnLimitReached(w, r)
			w.Header().Add("Content-Type", lmt.GetMessageContentType())
			w.WriteHeader(lmt.GetStatusCode())
			w.Write([]byte(lmt.GetMessage()))
			return
		}
		// There's no rate-limit error, serve the next handler.
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(middle)
}

func (mux *LServeMux) Handler(path string, qps float64, h http.Handler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	if path == "" {
		panic("http: invalid pattern")
	}

	if mux.m == nil {
		mux.m = make(map[*regexp.Regexp]http.Handler)
	}

	if h == nil {
		mux.m[regexp.MustCompile(path)] = LimitHandler(path, tollbooth.NewLimiter(qps, nil), mux.rp)
		return
	}

	mux.m[regexp.MustCompile(path)] = LimitHandler(path, tollbooth.NewLimiter(qps, nil), h)

}

func (mux *LServeMux) ListenAndServe() {
	http.ListenAndServe(mux.add, mux)
}

type LServeMux struct {
	mu  sync.RWMutex
	m   map[*regexp.Regexp]http.Handler
	rp  *httputil.ReverseProxy
	add string
}

func (mux *LServeMux) DefaultHandler() http.Handler {
	return mux.rp
}

func (mux *LServeMux) handler(r *http.Request) http.Handler {

	path := r.URL.Path
	if mux.m != nil {
		for k, v := range mux.m {
			if k.MatchString(path) {
				return v
			}
		}
	}

	return nil
}

func (mux *LServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := mux.handler(r)
	if h == nil {
		mux.rp.ServeHTTP(w, r)
	} else {
		h.ServeHTTP(w, r)
	}
}

// 反向代理到多个http服务逗号分隔 默认调度随机分配
// 例如 http://www.good1.com:8081,http://www.good2.com:8081
// Reverse proxy to multiple http services comma separated. Default dispatch random allocation
// For example http://www.good1.com:8081,http://www.good2.com:8081

func ReverseProxy(servlist string) (*httputil.ReverseProxy, error) {
	backserlist := strings.Split(servlist, ",")
	var urlbacklist []*url.URL
	for _, backser := range backserlist {
		u, err := url.Parse(backser)
		if err == nil {
			urlbacklist = append(urlbacklist, u)
		}
	}
	if len(urlbacklist) < 1 {
		return nil, errors.New("get back service ")
	}

	rand.Seed(time.Now().Unix())

	director := func(req *http.Request) {
		length := len(urlbacklist)
		target := urlbacklist[rand.Int()%length]
		targetQuery := target.RawQuery
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}

	return &httputil.ReverseProxy{Director: director}, nil
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
