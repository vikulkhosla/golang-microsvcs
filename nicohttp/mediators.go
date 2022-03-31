package nicohttp

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

func suspendMediator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if (atomic.LoadInt32(&builder.server.suspended) == 1) && !isBase(r.RequestURI) {
			http.Error(w, "Temporarily Suspended", http.StatusServiceUnavailable)
			return
		}
		next.ServeHTTP(w, r)
	})
}


func httpBasicAuthMediator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hdr := r.Header.Get("Authorization")
		if hdr == "" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		splits := strings.Fields(hdr)
		if !strings.EqualFold(splits[0], "BASIC") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		b64d, err := base64.StdEncoding.DecodeString(splits[1])
		if err != nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		user := strings.Split(string(b64d), ":")[0]
		r.Header.Add("X-AUTH-USER", user)
		next.ServeHTTP(w, r)
	})
}

func noAuthMediator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := "anonymous"
		r.Header.Add("X-AUTH-USER", user)
		next.ServeHTTP(w, r)
	})
}

// TBD
func hmacJWTMediator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := "hmac-jwt"
		r.Header.Add("X-AUTH-USER", user)
		next.ServeHTTP(w, r)
	})
}

// TBD
func rsaJWTMediator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := "rsa-jwt"
		r.Header.Add("X-AUTH-USER", user)
		next.ServeHTTP(w, r)
	})
}

// TBD
func ldapMediator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := "ldap-user"
		r.Header.Add("X-AUTH-USER", user)
		next.ServeHTTP(w, r)
	})
}

func loggingMediatorBefore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Header.Get("X-AUTH-USER")
		if user == "" {
			user = "unknown"
		}
		log.Println(user, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
		next.ServeHTTP(w, r)
	})
}

type statusResponseWriter struct {
	http.ResponseWriter
	status int
	length int
}

func (w *statusResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusResponseWriter) Write(b []byte) (int, error) {
	
	if w.status == 0 {
		w.status = 200
	}
	n, err := w.ResponseWriter.Write(b)
	w.length += n
	return n, err	
}

func memoryPostLoggingMediator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		sw := statusResponseWriter{ResponseWriter: w}
		next.ServeHTTP(&sw, r)
		user := r.Header.Get("X-AUTH-USER")
		if user == "" || user == "anonymous" {
			user = r.Header.Get("X-Goog-Authenticated-User-Email")
			if user == "" {
				user = r.Header.Get("X-CHARIOT-USER")
				if user == "" {
					user = "anonymous"
				}
			}
		}
		if r.RequestURI == "/healthz" || strings.HasPrefix(r.RequestURI, "/logs") {
			return
		}
		sc := sw.status
		cl := sw.length
		msg := fmt.Sprintf("Request: requestID=%s, user=%s, remoteAddr=%s, %s %s ; Response: status=%d, CT=%s, CL=%d",
			r.Header.Get("X-Request-ID"),user, r.RemoteAddr, r.Method, r.RequestURI, sc, w.Header().Get("Content-Type"), cl)
		builder.server.logChan <- msg
	})
}


func timeoutMediator(next http.Handler) http.Handler {
	return http.TimeoutHandler(next, builder.server.handlerTimeout, "timed out")
}



func tracingMediator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("%d", time.Now().UnixNano())
			//ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			r.Header.Set("X-Request-Id", requestID)
		}
		next.ServeHTTP(w, r)
	})
}


/*
func initRateLimiting() {
	store, err := memstore.New(65536)
    quota := throttled.RateQuota{throttled.PerMin(20), 5}
    rateLimiter, err := throttled.NewGCRARateLimiter(store, quota)
    if err != nil {
	    log.Fatal(err)
    }
    httpRateLimiter := throttled.HTTPRateLimiter{
	    RateLimiter: rateLimiter,
	    VaryBy:      &throttled.VaryBy{Path: true},
	}
	//th := throttled.Interval(throttled.PerSec(10), 1, &throttled.VaryBy{Path: true}, 50)
}

*/




