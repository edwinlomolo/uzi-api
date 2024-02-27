package handler

import "net/http"

func GetIp(r *http.Request) string {
	clientIp := r.Header.Get("X-FORWARDED-FOR")
	if clientIp != "" {
		return clientIp
	}

	return r.RemoteAddr
}
