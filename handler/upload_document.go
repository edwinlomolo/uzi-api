package handler

import (
	"encoding/json"
	"net/http"

	"github.com/edwinlomolo/uzi-api/internal"
)

func UploadDocument() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		maxSize := int64(6000000)
		upldr := internal.GetGCS()

		err := r.ParseMultipartForm(maxSize)
		if err != nil {
			log.WithError(err).Errorf("parse multipart form")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			log.WithError(err).Errorf("reading req form file value")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		imageUri, uploadErr := upldr.UploadCourierDocument(file, fileHeader)
		if uploadErr != nil {
			http.Error(w, uploadErr.Error(), http.StatusInternalServerError)
			return
		}

		res, marshalErr := json.Marshal(struct {
			ImageUri string `json:"imageUri"`
		}{ImageUri: imageUri})
		if marshalErr != nil {
			log.WithError(marshalErr).Errorf("marshal upload res")
			http.Error(w, marshalErr.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	})
}
