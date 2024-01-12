package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/3dw1nM0535/uzi-api/aws"
	"github.com/3dw1nM0535/uzi-api/logger"
	"github.com/3dw1nM0535/uzi-api/model"
)

func UploadDocument() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		maxSize := int64(6000000)
		logger := logger.GetLogger()
		s3Service := aws.GetS3Service()

		err := r.ParseMultipartForm(maxSize)
		if err != nil {
			uziErr := model.UziErr{Err: errors.New("FileTooLarge").Error(), Message: "uploadimage", Code: http.StatusBadRequest}
			logger.Errorf(uziErr.Error())
			http.Error(w, uziErr.Error(), uziErr.Code)
			return
		}

		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			uziErr := model.UziErr{Err: errors.New("ExpectedFile").Error(), Message: "nofile", Code: http.StatusBadRequest}
			logger.Errorf(uziErr.Error())
			http.Error(w, uziErr.Error(), uziErr.Code)
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
			uziErr := model.UziErr{Err: errors.New("UploadImageMarshal").Error(), Message: "marshal", Code: http.StatusBadRequest}
			logger.Errorf(uziErr.Error())
			http.Error(w, uziErr.Error(), uziErr.Code)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	})
}
