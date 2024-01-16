package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/3dw1nM0535/uzi-api/model"
	"github.com/3dw1nM0535/uzi-api/pkg/logger"
	"github.com/3dw1nM0535/uzi-api/pkg/util"
	"github.com/3dw1nM0535/uzi-api/services/ipinfo"
)

func Ipinfo() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := logger.GetLogger()
		ip := util.GetIp(r)

		ipinfo, err := ipinfo.GetIpinfoService().GetIpinfo(ip)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		res, jsonErr := json.Marshal(ipinfo)
		if jsonErr != nil {
			uziErr := model.UziErr{Err: errors.New("JsonMarshalErr").Error(), Message: "marshal", Code: http.StatusInternalServerError}
			logger.Errorf(uziErr.Error())
			http.Error(w, uziErr.Error(), uziErr.Code)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	})
}
