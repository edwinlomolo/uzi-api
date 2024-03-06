package handler

import (
	"encoding/json"
	"net/http"

	"github.com/edwinlomolo/uzi-api/aws"
	"github.com/edwinlomolo/uzi-api/logger"
)

func UploadDocument() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		maxSize := int64(6000000)
		logger := logger.New()
		s3Service := aws.GetAws()

		err := r.ParseMultipartForm(maxSize)
		if err != nil {
			logger.WithError(err).Errorf("parse multipart form")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			logger.WithError(err).Errorf("reading req form file value")
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
			logger.WithError(marshalErr).Errorf("marshal upload res")
			http.Error(w, marshalErr.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	})
}
