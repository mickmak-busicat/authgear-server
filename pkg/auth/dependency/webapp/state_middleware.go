package webapp

import (
	"net/http"

	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/intl"
)

type StateMiddlewareStates interface {
	Get(instanceID string) (*State, error)
}

type StateMiddleware struct {
	States StateMiddlewareStates
}

func (m *StateMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		q := r.URL.Query()
		sid := q.Get("x_sid")

		if sid == "" {
			next.ServeHTTP(w, r)
			return
		}

		state, err := m.States.Get(sid)
		if err != nil {
			u := *r.URL
			q := u.Query()
			RemoveX(q)
			u.RawQuery = q.Encode()
			u.Path = "/"

			http.Redirect(w, r, httputil.HostRelative(&u).String(), http.StatusFound)
			return
		}

		// Restore UI locales from state if necessary
		if state.UILocales != "" {
			tags := intl.ParseUILocales(state.UILocales)
			ctx := intl.WithPreferredLanguageTags(r.Context(), tags)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}
