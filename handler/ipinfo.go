package handler

import (
	"encoding/json"
	"net/http"

	"github.com/3dw1nM0535/uzi-api/services"
	"github.com/sirupsen/logrus"
)

func Ipinfo() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := ctx.Value("logger").(*logrus.Logger)
		ip := GetIp(r)

		ipinfo, err := ctx.Value("ipinfoService").(services.IpInfo).GetIpinfo(ip)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		res, err := json.Marshal(ipinfo)
		if err != nil {
			logger.Errorf("%s-%v", "JsonMarshalErr", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	})
}
