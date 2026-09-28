package handler

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/ulibaysya/muzalea/internal/modules/audio/handler/dto"
	pkghttp "github.com/ulibaysya/muzalea/pkg/http"
	pkgslog "github.com/ulibaysya/muzalea/pkg/slog"
)

// TODO non-necessary refactor
func (h Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	mr, err := r.MultipartReader()
	if err != nil {
		h.logger.LogAttrs(ctx, slog.LevelWarn, "obtaining multipart reader", pkgslog.Error("error", err))
		pkghttp.ErrorJSON400(w)
		return
	}

	metaPart, err := mr.NextPart()
	if err != nil {
		h.logger.LogAttrs(ctx, slog.LevelWarn, "obtaining meta multipart part", pkgslog.Error("error", err))
		pkghttp.ErrorJSON400(w)
		return
	}
	defer func() {
		if err := metaPart.Close(); err != nil {
			h.logger.LogAttrs(ctx, slog.LevelWarn, "closing meta multipart part", pkgslog.Error("error", err))
		}
	}()
	if formName := metaPart.FormName(); formName != "meta" {
		h.logger.LogAttrs(ctx, slog.LevelWarn, "expected meta part", slog.String("formName", formName))
		pkghttp.ErrorJSON400(w)
		return
	}
	metaRequest := dto.CreationRequest{}
	if err = json.UnmarshalRead(metaPart, &metaRequest); err != nil {
		h.logger.LogAttrs(ctx, slog.LevelWarn, "parsing meta", pkgslog.Error("error", err))
		pkghttp.ErrorJSON400(w)
		return
	}
	if metaRequest.IsZero() {
		h.logger.LogAttrs(ctx, slog.LevelWarn, "some meta fields are omitted")
		pkghttp.ErrorJSON400(w)
		return
	}

	dataPart, err := mr.NextPart()
	if err != nil {
		h.logger.LogAttrs(ctx, slog.LevelWarn, "obtaining data multipart part", pkgslog.Error("error", err))
		pkghttp.ErrorJSON400(w)
		return
	}
	defer func() {
		if errClose := dataPart.Close(); errClose != nil {
			h.logger.LogAttrs(ctx, slog.LevelWarn, "closing data multipart part", pkgslog.Error("error", errClose))
		}
	}()
	if formName := dataPart.FormName(); formName != "data" {
		h.logger.LogAttrs(ctx, slog.LevelWarn, "expected data part", slog.String("formName", formName))
		pkghttp.ErrorJSON400(w)
		return
	}
	dataBuffer := bytes.NewBuffer(nil)
	const maxDataSize = 20 << 20
	n, err := io.Copy(dataBuffer, io.LimitReader(dataPart, maxDataSize+1))
	if err != nil {
		h.logger.LogAttrs(ctx, slog.LevelWarn, "copying data multipart part to memory buffer", pkgslog.Error("error", err))
		pkghttp.ErrorJSON400(w)
		return
	}
	if n > maxDataSize {
		h.logger.LogAttrs(ctx, slog.LevelWarn, "data multipart part exceeded limit", slog.Int("maxDataSize", maxDataSize))
		pkghttp.ErrorJSON413(w)
		return
	}

	if _, err = mr.NextPart(); err == nil || !errors.Is(err, io.EOF) {
		if err != nil {
			h.logger.LogAttrs(ctx, slog.LevelWarn, "error when checking for unexpected additional multipart part", pkgslog.Error("error", err))
		} else {
			h.logger.LogAttrs(ctx, slog.LevelWarn, "unexpected additional multipart part")
		}
		pkghttp.ErrorJSON400(w)
		return
	}
	
	audioID, err := h.service.SaveAudio(context.WithoutCancel(ctx), *metaRequest.Name, *metaRequest.Author, metaRequest.Tags, dataBuffer)
	if err != nil {
		h.logger.LogAttrs(ctx, slog.LevelWarn, "saving audio", pkgslog.Error("error", err))
		pkghttp.ErrorJSON400(w)
		return
	}

	w.Header().Set(pkghttp.HeaderLocation, fmt.Sprintf("/audio/%d", audioID))
	w.WriteHeader(http.StatusCreated)
}
