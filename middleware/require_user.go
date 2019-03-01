package middleware

import (
	"net/http"
	"strings"

	"lenslocked.com/context"
	"lenslocked.com/models"
)

type User struct {
	models.UserService
}

func (mw *User) Apply(next http.Handler) http.HandlerFunc {
	return mw.ApplyFn(next.ServeHTTP)
}
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
		// fmt.Println(r.URL.Path)
		cookie, err := r.Cookie("remember_token")
		if err != nil {
			next(w, r)
			return
		}
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
