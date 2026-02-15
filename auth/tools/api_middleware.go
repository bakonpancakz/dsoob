package tools

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SessionData struct {
	SessionID int64 // Relevant Session ID
	UserID    int64 // Relevant User ID
	Elevated  bool  // Relevant Session Elevated?
}
type RatelimitOptions struct {
	Period time.Duration // Reset Period
	Limit  int64         // Maximum Amount of Requests
}

type RatelimitEntry struct {
	Usage     int64
	ExpiresAt int64
}

// Protect Server against Abuse by Limiting the amount of incoming bytes
func NewBodyLimit(limit int64) MiddlewareFunc {
	return func(w http.ResponseWriter, r *http.Request) bool {
		r.Body = http.MaxBytesReader(w, r.Body, limit)
		if r.ContentLength > limit {
			SendClientError(w, r, ERROR_BODY_TOO_LARGE)
			return false
		}
		return true
	}
}

// Protect Server against Abuse by Limiting the amount of incoming requests
func NewRatelimit(o *RatelimitOptions) MiddlewareFunc {

	var mtx sync.Mutex
	var data = make(map[string]*RatelimitEntry, 1024)

	// Cleanup Internval
	interval := time.NewTicker(time.Minute)
	go func() {
		for range interval.C {
			mtx.Lock()
			now := time.Now().Unix()
			for k, v := range data {
				if now > v.ExpiresAt {
					delete(data, k)
				}
			}
			mtx.Unlock()
		}
	}()

	return func(w http.ResponseWriter, r *http.Request) bool {
		key := (r.Method + r.URL.Path + GetRemoteIP(r))
		now := time.Now().Unix()

		mtx.Lock()
		defer mtx.Unlock()

		// Fetch Key
		e, ok := data[key]
		if ok && now > e.ExpiresAt {
			delete(data, key)
			ok = false
		} else if !ok {
			e = &RatelimitEntry{
				Usage:     1,
				ExpiresAt: now + int64(o.Period.Seconds()),
			}
			data[key] = e
		} else {
			e.Usage++
		}

		// Append Headers
		var (
			ttl       = max(e.ExpiresAt-now, 0)
			remaining = max(o.Limit-e.Usage, 0)
		)
		w.Header().Set("X-Ratelimit-Remaining", strconv.FormatInt(remaining, 10))
		w.Header().Set("X-Ratelimit-Reset", strconv.FormatInt(ttl, 10))
		w.Header().Set("X-Ratelimit-Limit", strconv.FormatInt(o.Limit, 10))

		// Enforce Limits
		if e.Usage > o.Limit {
			SendClientError(w, r, ERROR_GENERIC_RATELIMIT)
			return false
		}

		return true
	}
}

// Retrieve User or Application Session from Request
func UseSession(w http.ResponseWriter, r *http.Request) bool {

	h := strings.TrimSpace(r.Header.Get("Authorization"))
	switch {

	// User Prefix
	case strings.HasPrefix(h, TOKEN_PREFIX_USER):
		ctx, cancel := NewContext()
		defer cancel()

		// Retrieve User Session
		var session SessionData
		var sessionElevatedUntil time.Time

		err := Database.QueryRowContext(ctx,
			"SELECT id, user_id, elevated_until FROM user_session WHERE token = $1",
			strings.TrimPrefix(h, TOKEN_PREFIX_USER),
		).Scan(
			&session.SessionID,
			&session.UserID,
			&sessionElevatedUntil,
		)
		if errors.Is(err, sql.ErrNoRows) {
			SendClientError(w, r, ERROR_GENERIC_UNAUTHORIZED)
			return false
		}
		if err != nil {
			SendServerError(w, r, err)
			return false
		}

		// Additional Checks
		if time.Now().Before(sessionElevatedUntil) {
			session.Elevated = true
		}

		// Apply Session to Request Context
		ctxWithSession := context.WithValue(r.Context(), SESSION_KEY, &session)
		*r = *r.WithContext(ctxWithSession)

	// Unknown Prefix
	default:
		SendClientError(w, r, ERROR_GENERIC_UNAUTHORIZED)
		return false

	}

	return true
}
