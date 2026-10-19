package handler

import (
	"compress/gzip"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func Gzip() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if hasEncoding(ctx.GetHeader("Content-Encoding"), "gzip") {
			reader, err := gzip.NewReader(ctx.Request.Body)
			if err != nil {
				ctx.AbortWithStatus(http.StatusBadRequest)
				return
			}
			ctx.Request.Body = &gzipReadCloser{
				Reader: reader,
				body:   ctx.Request.Body,
			}
			ctx.Request.Header.Del("Content-Encoding")
			ctx.Request.ContentLength = -1
		}

		if !acceptsGzip(ctx.GetHeader("Accept-Encoding")) {
			ctx.Next()
			return
		}

		writer := &gzipResponseWriter{ResponseWriter: ctx.Writer}
		ctx.Writer = writer
		ctx.Next()
		if writer.writer != nil {
			_ = writer.writer.Close()
		}
	}
}

type gzipReadCloser struct {
	*gzip.Reader
	body io.Closer
}

func (r *gzipReadCloser) Close() error {
	readerErr := r.Reader.Close()
	bodyErr := r.body.Close()
	if readerErr != nil {
		return readerErr
	}
	return bodyErr
}

type gzipResponseWriter struct {
	gin.ResponseWriter
	writer             *gzip.Writer
	compressionChecked bool
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	w.startCompression()
	if w.writer == nil {
		return w.ResponseWriter.Write(data)
	}
	return w.writer.Write(data)
}

func (w *gzipResponseWriter) WriteString(data string) (int, error) {
	return w.Write([]byte(data))
}

func (w *gzipResponseWriter) Flush() {
	w.startCompression()
	if w.writer != nil {
		_ = w.writer.Flush()
	}
	w.ResponseWriter.Flush()
}

func (w *gzipResponseWriter) startCompression() {
	if w.compressionChecked {
		return
	}
	w.compressionChecked = true

	if !isCompressible(w.Header().Get("Content-Type")) {
		return
	}

	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Add("Vary", "Accept-Encoding")
	w.Header().Del("Content-Length")
	w.writer = gzip.NewWriter(w.ResponseWriter)
}

func isCompressible(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	return mediaType == "application/json" || mediaType == "text/html"
}

func hasEncoding(header, encoding string) bool {
	for value := range strings.SplitSeq(header, ",") {
		if strings.EqualFold(strings.TrimSpace(value), encoding) {
			return true
		}
	}
	return false
}

func acceptsGzip(header string) bool {
	for value := range strings.SplitSeq(header, ",") {
		parts := strings.Split(strings.TrimSpace(value), ";")
		if !strings.EqualFold(strings.TrimSpace(parts[0]), "gzip") {
			continue
		}

		accepted := true
		for _, parameter := range parts[1:] {
			name, value, ok := strings.Cut(strings.TrimSpace(parameter), "=")
			if !ok || !strings.EqualFold(name, "q") {
				continue
			}
			quality, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
			accepted = err == nil && quality > 0
		}
		return accepted
	}
	return false
}
