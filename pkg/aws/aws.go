package aws

import "mime/multipart"

type Aws interface {
	UploadImage(multipart.File, *multipart.FileHeader) (string, error)
}
