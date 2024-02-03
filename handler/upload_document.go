package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/3dw1nM0535/uzi-api/internal/aws"
	"github.com/3dw1nM0535/uzi-api/internal/logger"
)

func UploadDocument() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		maxSize := int64(6000000)
		logger := logger.Logger
		s3Service := aws.S3

		err := r.ParseMultipartForm(maxSize)
		if err != nil {
			uziErr := fmt.Errorf("%s:%v", "parsing form multipart", err)
			logger.Errorf(uziErr.Error())
			http.Error(w, uziErr.Error(), http.StatusBadRequest)
			return
		}

		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			uziErr := fmt.Errorf("%s:%v", "reading form file", err)
			logger.Errorf(uziErr.Error())
			http.Error(w, uziErr.Error(), http.StatusBadRequest)
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
			uziErr := fmt.Errorf("%s:%v", "marshal upload res", marshalErr)
			logger.Errorf(uziErr.Error())
			http.Error(w, uziErr.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	})
}
