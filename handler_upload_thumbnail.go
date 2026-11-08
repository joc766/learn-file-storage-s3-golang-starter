package main

import (
	"fmt"
	"io"
	"net/http"

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

	const maxMemory = 10 << 20
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse multipart form", err)
		return
	}

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't read form file", err)
		return
	}
	defer file.Close()
	mediaType := header.Header.Get("Content-Type")

	bytes, err := io.ReadAll(file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't read file", err)
		return
	}

	video, err := cfg.db.GetVideo(videoID)
	if userID != video.UserID {
		respondWithError(w, http.StatusUnauthorized, "Authenticated user does not match video owner", nil)
		return
	}

	newThumbnail := thumbnail{
		data:      bytes,
		mediaType: mediaType,
	}

	videoThumbnails[videoID] = newThumbnail
	// why is ThumbnailURL a pointer to a string?
	newThumnailUrl := fmt.Sprintf("http://localhost:%v/api/thumbnails/%v", 8091, videoID)
	video.ThumbnailURL = &newThumnailUrl
	if err := cfg.db.UpdateVideo(video); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not update video in db", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
