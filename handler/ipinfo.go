package handler

import (
	"encoding/json"
	"net/http"

	"github.com/edwinlomolo/uzi-api/internal"
	"github.com/edwinlomolo/uzi-api/services"
)

var (
	log           = internal.GetLogger()
	ipinfoService = services.GetIpinfoService()
)

func Ipinfo() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.Context().Value("ip").(string)

		ipinfo, err := ipinfoService.GetIpinfo(ip)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		res, jsonErr := json.Marshal(ipinfo)
		if jsonErr != nil {
			log.WithError(jsonErr).Errorf("marshal ipinfo res")
			http.Error(w, jsonErr.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	})
}
