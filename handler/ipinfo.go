package handler

import (
	"encoding/json"
	"net/http"

	"github.com/edwinlomolo/uzi-api/ipinfo"
	"github.com/edwinlomolo/uzi-api/logger"
)

func Ipinfo(ipinfoService ipinfo.IpInfoService) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logger.New()
		ip := GetIp(r)

		ipinfo, err := ipinfoService.GetIpinfo(ip)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		res, jsonErr := json.Marshal(ipinfo)
		if jsonErr != nil {
			logger.WithError(jsonErr).Errorf("marshal ipinfo res")
			http.Error(w, jsonErr.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	})
}
