package handler

import (
	"encoding/json"
	"net/http"

	"github.com/3dw1nM0535/uzi-api/logger"
	"github.com/3dw1nM0535/uzi-api/pkg/aws"
)

func UploadDocument() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		maxSize := int64(6000000)
		logger := logger.GetLogger()
		s3Service := aws.GetAwsService()

		logger.Infoln(r.FormValue("reason"))
		err := r.ParseMultipartForm(maxSize)
		if err != nil {
			errMsg := "FileTooLarge"
			logger.Errorf("%s: %v", errMsg, err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			errMsg := "ExpectedFile"
			logger.Errorf("%s: %v", errMsg, err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		imageUri, uploadErr := s3Service.UploadImage(file, fileHeader)
		if uploadErr != nil {
			http.Error(w, uploadErr.Error(), http.StatusInternalServerError)
			return
		}

		res, marshalErr := json.Marshal(struct {
			ImageUri string `json:"imageUri"`
		}{ImageUri: imageUri})
		if marshalErr != nil {
			errMsg := "UploadImageMarshalResErr"
			logger.Errorf("%s: %v", errMsg, marshalErr.Error())
			http.Error(w, marshalErr.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	})
}
