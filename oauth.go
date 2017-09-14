package main

import (
	"context"
	"net/http"

	"github.com/ChimeraCoder/anaconda"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/tsuna/gohbase/hrpc"
)

// OAuthInit initializes OAuth 2.0 handshake
func OAuthInit(w http.ResponseWriter, r *http.Request) {
	session, error := store.Get(r, "twister0")
	if error != nil {
		respond(w, nil, error.Error(), http.StatusInternalServerError)
		return
	}
	url, credentials, error := anaconda.AuthorizationURL("http://localhost:8081/oauth/callback")
	if error != nil {
		respond(w, nil, error.Error(), http.StatusInternalServerError)
		return
	}
	session.Values["credentials"] = credentials
	session.Save(r, w)
	http.Redirect(w, r, url, http.StatusFound)
}

// OAuthCallback completes OAuth 2.0 handshake
func OAuthCallback(w http.ResponseWriter, r *http.Request) {
	session, error := store.Get(r, "twister0")
	if error != nil {
		respond(w, nil, error.Error(), http.StatusInternalServerError)
		return
	}
	val := session.Values["credentials"]
	credentials, fucked := val.(*oauth.Credentials)
	if fucked == false {
		respond(w, nil, "Some shit went wrong.", http.StatusInternalServerError)
		return
	}
	credentials, _, error = anaconda.GetCredentials(credentials, r.URL.Query().Get("oauth_verifier"))
	if error != nil {
		respond(w, nil, error.Error(), http.StatusInternalServerError)
		return
	}
	api := anaconda.NewTwitterApi(credentials.Token, credentials.Secret)
	result, error := api.GetSelf(nil)
	if error != nil {
		respond(w, nil, error.Error(), http.StatusInternalServerError)
		return
	}
	values := map[string]map[string][]byte{"user": map[string][]byte{"token": []byte(credentials.Token), "tokenSecret": []byte(credentials.Secret)}}
	putRequest, error := hrpc.NewPutStr(context.Background(), "twister0", result.IdStr, values)
	_, error = client.Put(putRequest)
	if error != nil {
		respond(w, nil, error.Error(), http.StatusInternalServerError)
		return
	}
	delete(session.Values, "credentials")
	session.Values["idStr"] = result.IdStr
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

// OAuthSelf returns authenticated self
func OAuthSelf(w http.ResponseWriter, r *http.Request) {
	session, error := store.Get(r, "twister0")
	if error != nil {
		respond(w, nil, error.Error(), http.StatusInternalServerError)
		return
	}
	// val := session.Values["credentials"]
	// credentials, fucked := val.(*oauth.Credentials)
	// if fucked == false {
	// 	respond(w, nil, "Some shit went wrong.", http.StatusInternalServerError)
	// 	return
	// }
	// api := anaconda.NewTwitterApi(credentials.Token, credentials.Secret)
	// _, error = api.GetSelf(url.Values{"skip_status": {string("true")}})
	// // result, error := api.GetSelf(url.Values{"skip_status": {string("true")}})
	// if error != nil {
	// 	respond(w, nil, error.Error(), http.StatusInternalServerError)
	// 	return
	// }
	respond(w, session.Values["idStr"], nil, http.StatusOK)
	// respond(w, result, nil, http.StatusOK)
}
