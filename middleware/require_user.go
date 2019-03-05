package middleware

import (
	"net/http"
	"strings"
	"time"

	"lenslocked.com/context"
	"lenslocked.com/models"
)

type User struct {
	models.UserService
}

func (mw *User) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFn(next.ServeHTTP)
}

// ApplyFn will apply middleware to all users.  First it checks if the user is
// fetching static images or assets which are permitted everywhere and any
// user can load them without being looked up in the db.  Then cookies are
// checked for all other requests by getting the remember_token value.  If the
// cookie is expired it will redirect the next handler will be run without a user
// being set otherwise the cookies expiration date will be set to an hour from
// the time they try to access a page on the server.  User is then qu
func (mw *User) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// don't require user middleware if user is requesting
		// static asset or image so no lookup of user is required
		if strings.HasPrefix(path, "/assets/") ||
			strings.HasPrefix(path, "/images/") {
			next(w, r)
			return
		}
		cookie, err := r.Cookie("remember_token")
		if err != nil {
			next(w, r)
			return
		}
		cookie.Expires = time.Now().Add(time.Hour)
		user, err := mw.ByRemember(cookie.Value)
		if err != nil {
			next(w, r)
			return
		}
		ctx := r.Context()
		ctx = context.WithUser(ctx, user)
		r = r.WithContext(ctx)
		next(w, r)
	})
}

// Owner assumes that User middleware has already
// been run otherwise it will not work correctly.
// Redirects user to login page if they are viewing something
// they do not have access to as they need to be the owner of the
// content they wish to view
type Owner struct {
	User
}

// Apply assumes that User middleware has already
// been run otherwise it will not work correctly
func (mw *Owner) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFn(next.ServeHTTP)
}

// ApplyFn assumes that User middleware has already
// been run otherwise it will not work correctly
func (mw *Owner) ApplyFn(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := context.User(r.Context())
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next(w, r)
	})
}
