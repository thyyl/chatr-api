package uploader

import (
	"context"
	base64 "encoding/base64"
	"io"
	"net/http"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gin-gonic/gin"
	"github.com/thyyl/chatr/pkg/common"
)

func (s *HttpServer) UploadFiles(ctx *gin.Context) {
	channelId, ok := ctx.Request.Context().Value(common.ChannelKey).(uint64)
	if !ok {
		common.Response(ctx, http.StatusUnauthorized, common.ErrorUnauthorized)
	}

	if err := ctx.Request.ParseMultipartForm(s.maxMemory); err != nil {
		s.logger.Error("failed to parse multipart form" + err.Error())
		common.Response(ctx, http.StatusInternalServerError, common.ErrorServer)
		return
	}

	form, err := ctx.MultipartForm()
	if err != nil {
		s.logger.Error("failed to get multipart form" + err.Error())
		common.Response(ctx, http.StatusBadRequest, common.ErrorReceiveFile)
		return
	}

	fileHeaders := form.File["files"]
	var uploadedFiles []UploadedFileDto

	for _, fileHeader := range fileHeaders {
		file, err := fileHeader.Open()
		if err != nil {
			s.logger.Error("failed to open multipart file" + err.Error())
			common.Response(ctx, http.StatusBadRequest, common.ErrorReceiveFile)
			return
		}

		extension := filepath.Ext(fileHeader.Filename)
		newFileName := newObjectKey(channelId, extension)
		if err := s.putFileToS3(ctx.Request.Context(), s.s3Bucket, newFileName, file); err != nil {
			s.logger.Error("failed to upload file" + err.Error())
			common.Response(ctx, http.StatusInternalServerError, common.ErrorServer)
			return
		}

		uploadedFiles = append(uploadedFiles, UploadedFileDto{
			Name: fileHeader.Filename,
			Url:  joinStrings(s.s3Endpoint, "/", s.s3Bucket, "/", newFileName),
		})
	}

	ctx.JSON(http.StatusCreated, &UploadedFilesDto{
		UploadedFiles: uploadedFiles,
	})
}

func (s *HttpServer) GetPresignedUpload(ctx *gin.Context) {
	channelId, ok := ctx.Request.Context().Value(common.ChannelKey).(uint64)
	if !ok {
		common.Response(ctx, http.StatusUnauthorized, common.ErrorUnauthorized)
		return
	}

	var request GetPresignedUrlRequest
	if err := ctx.ShouldBindQuery(&request); err != nil {
		common.Response(ctx, http.StatusBadRequest, common.ErrorInvalidParam)
		return
	}

	objectKey := newObjectKey(channelId, common.Join(".", request.Extension))
	response, err := s.presigner.PutObject(ctx.Request.Context(), s.s3Bucket, objectKey)
	if err != nil {
		s.logger.Error("failed to get upload presigned url" + err.Error())
		common.Response(ctx, http.StatusInternalServerError, common.ErrorServer)
		return
	}

	ctx.JSON(http.StatusOK, &PresignedUpload{
		Url:       response.URL,
		ObjectKey: objectKey,
	})
}

func (s *HttpServer) GetPresignedDownload(ctx *gin.Context) {
	channelId, ok := ctx.Request.Context().Value(common.ChannelKey).(uint64)
	if !ok {
		common.Response(ctx, http.StatusUnauthorized, common.ErrorUnauthorized)
		return
	}

	var request GetPresignedDownloadRequest
	if err := ctx.ShouldBindQuery(&request); err != nil {
		common.Response(ctx, http.StatusBadRequest, common.ErrorInvalidParam)
		return
	}

	objectKeyByte, err := base64.URLEncoding.DecodeString(request.ObjectKeyBase64)
	if err != nil {
		common.Response(ctx, http.StatusBadRequest, common.ErrorInvalidParam)
		return
	}

	objectKey := byteSlice2String(objectKeyByte)

	targetChannelId, err := getChannelIdFromObjectKey(objectKey)
	if err != nil {
		common.Response(ctx, http.StatusBadRequest, common.ErrorInvalidParam)
		return
	}

	if targetChannelId != channelId {
		common.Response(ctx, http.StatusUnauthorized, common.ErrorUnauthorized)
		return
	}

	response, err := s.presigner.GetObject(ctx.Request.Context(), s.s3Bucket, objectKey)
	if err != nil {
		s.logger.Error("failed to get download presigned url" + err.Error())
		common.Response(ctx, http.StatusInternalServerError, common.ErrorServer)
		return
	}

	ctx.JSON(http.StatusOK, &PresignedDownload{
		Url: response.URL,
	})

}

func (r *HttpServer) putFileToS3(ctx context.Context, bucket, fileName string, file io.Reader) error {
	_, err := r.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileName),
		ACL:    types.ObjectCannedACLPublicRead,
		Body:   file,
	})
	if err != nil {
		return err
	}
	return nil
}
