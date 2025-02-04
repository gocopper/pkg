package cauth

import (
	"net/http"
)

func GetLogoutHTTPCookies() []http.Cookie {
	cookies := getHTTPCookies("", "")
	for i := range cookies {
		cookies[i].MaxAge = -1
	}
	return cookies
}

func getHTTPCookies(sessionUUID, token string) []http.Cookie {
	return []http.Cookie{
		{
			Name:     "SessionUUID",
			Value:    sessionUUID,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   86_400, // 1 day expiration
		},
		{
			Name:     "SessionToken",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			MaxAge:   86_400, // 1 day expiration
		},
	}
}
