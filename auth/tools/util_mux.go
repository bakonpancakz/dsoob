package tools

import "net/http"

type MiddlewareFunc func(w http.ResponseWriter, r *http.Request) bool

// Apply Middleware before Processing Request
func Chain(h http.HandlerFunc, mw ...MiddlewareFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		for i := 0; i < len(mw); i++ {
			if !mw[i](w, r) {
				return
			}
		}
		h(w, r)
	}
}

type MethodHandler map[string]http.HandlerFunc

func (mh MethodHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler, ok := mh[r.Method]; ok {
		handler(w, r)
	} else {
		SendClientError(w, r, ERROR_GENERIC_METHOD_NOT_ALLOWED)
	}
}
