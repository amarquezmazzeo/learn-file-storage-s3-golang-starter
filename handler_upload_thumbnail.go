package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here

	const maxMemory = 10 << 20
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse multipart form data", err)
		return
	}

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse uploaded file data", err)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	// rawFile, err := io.ReadAll(file)
	// if err != nil {
	// 	respondWithError(w, http.StatusInternalServerError, "Couldn't read uploaded file data", err)
	// 	return
	// }

	videoMetadata, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't locate videoID in DB", err)
		return
	}
	if videoMetadata.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "VideoID does not belong to user", errors.New("video.UserID != authenticated UserID"))
		return
	}

	// tn := thumbnail{data: rawFile, mediaType: contentType}
	// videoThumbnails[videoID] = tn

	fileName := fmt.Sprintf("%s.%s", videoIDString, contentType[len(contentType)-3:])
	filePath := filepath.Join(cfg.assetsRoot, fileName)
	outputFile, err := os.Create(filePath)
	log.Printf("file path: %s\n", filePath)
	defer outputFile.Close()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create thumbnail image file", err)
		return
	}
	_, err = io.Copy(outputFile, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't write thumbnail image file", err)
		return
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}

	thumbnailURL := fmt.Sprintf("http://localhost:%s/assets/%s", port, fileName)
	videoMetadata.ThumbnailURL = &thumbnailURL
	err = cfg.db.UpdateVideo(videoMetadata)

	respondWithJSON(w, http.StatusOK, videoMetadata)
}
