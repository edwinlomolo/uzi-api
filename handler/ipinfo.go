package handler

import (
	"encoding/json"
	"net/http"

	"github.com/edwinlomolo/uzi-api/controllers"
	"github.com/edwinlomolo/uzi-api/internal"
)

func Ipinfo() http.HandlerFunc {
	ipinfoController := controllers.GetIpinfoController()
	log := internal.GetLogger()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.Context().Value("ip").(string)

		ipinfo, err := ipinfoController.GetIpinfo(ip)
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
