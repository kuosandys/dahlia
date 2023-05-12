package main

import (
	"fmt"
	"net/http"
)

func (a *application) handleRedirect(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not in returned URL params.", http.StatusOK)
		return
	}

	refreshToken, err := a.dropboxClient.GetRefreshToken(code, REDIRECT_URI)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting refresh token: %s", err), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(fmt.Sprintf("Refresh token: %s", refreshToken)))
}
