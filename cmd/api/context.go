package main

import (
	"context"
	"movies.cosmasgithinji.net/internal/data"
	"net/http"
)

type contextKey string

const userContextkey = contextKey("user")

func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextkey, user)
	return r.WithContext(ctx)
}

func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextkey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}
