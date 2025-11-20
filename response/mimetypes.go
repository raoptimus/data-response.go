/**
 * This file is part of the raoptimus/data-response.go library
 *
 * @copyright Copyright (c) Evgeniy Urvantsev
 * @license https://github.com/raoptimus/data-response.go/blob/master/LICENSE.md
 * @link https://github.com/raoptimus/data-response.go
 */

package response

// MimeType represents a media type (MIME type) for HTTP Content-Type header.
type MimeType string

// Common MIME types for various file and content types.
const (

	// Text MIME types

	MimeTypePlainText  MimeType = "text/plain; charset=utf-8"
	MimeTypeHTML       MimeType = "text/html; charset=utf-8"
	MimeTypeCSS        MimeType = "text/css; charset=utf-8"
	MimeTypeJavaScript MimeType = "application/javascript; charset=utf-8"
	MimeTypeXML        MimeType = "application/xml; charset=utf-8"
	MimeTypeCSV        MimeType = "text/csv; charset=utf-8"

	// Application MIME types

	MimeTypeJSON          MimeType = "application/json; charset=utf-8"
	MimeTypeFormData      MimeType = "application/x-www-form-urlencoded"
	MimeTypeMultipartForm MimeType = "multipart/form-data"
	MimeTypePDF           MimeType = "application/pdf"
	MimeTypeZip           MimeType = "application/zip"
	MimeTypeTar           MimeType = "application/x-tar"
	MimeTypeGzip          MimeType = "application/gzip"
	MimeTypeRar           MimeType = "application/x-rar-compressed"
	MimeType7z            MimeType = "application/x-7z-compressed"
	MimeTypeWasm          MimeType = "application/wasm"
	MimeTypeOctetStream   MimeType = "application/octet-stream"

	// Image MIME types

	MimeTypeJPEG MimeType = "image/jpeg"
	MimeTypePNG  MimeType = "image/png"
	MimeTypeGIF  MimeType = "image/gif"
	MimeTypeWebP MimeType = "image/webp"
	MimeTypeSVG  MimeType = "image/svg+xml"
	MimeTypeICO  MimeType = "image/x-icon"
	MimeTypeBMP  MimeType = "image/bmp"
	MimeTypeTIFF MimeType = "image/tiff"

	// Audio MIME types

	MimeTypeMP3  MimeType = "audio/mpeg"
	MimeTypeWAV  MimeType = "audio/wav"
	MimeTypeOgg  MimeType = "audio/ogg"
	MimeTypeAAC  MimeType = "audio/aac"
	MimeTypeWebM MimeType = "audio/webm"
	MimeTypeFlac MimeType = "audio/flac"

	// Video MIME types

	MimeTypeMP4       MimeType = "video/mp4"
	MimeTypeWebMVideo MimeType = "video/webm"
	MimeTypeMPEG      MimeType = "video/mpeg"
	MimeTypeAVI       MimeType = "video/x-msvideo"
	MimeTypeQuickTime MimeType = "video/quicktime"
	MimeTypeMatroska  MimeType = "video/x-matroska"
	MimeTypeFlv       MimeType = "video/x-flv"

	// API MIME types

	MimeTypeJSONAPI     MimeType = "application/vnd.api+json"
	MimeTypeHAL         MimeType = "application/hal+json"
	MimeTypeProblemJSON MimeType = "application/problem+json"
)

// String returns the string representation of MimeType.
func (m MimeType) String() string {
	return string(m)
}

// MimeTypeMapping maps file extensions to MIME types.
// It can be used to automatically detect the correct MIME type based on file extension.
var MimeTypeMapping = map[string]MimeType{
	// Text extensions
	".txt":  MimeTypePlainText,
	".html": MimeTypeHTML,
	".htm":  MimeTypeHTML,
	".css":  MimeTypeCSS,
	".js":   MimeTypeJavaScript,
	".xml":  MimeTypeXML,
	".csv":  MimeTypeCSV,

	// Application extensions
	".json": MimeTypeJSON,
	".pdf":  MimeTypePDF,
	".zip":  MimeTypeZip,
	".tar":  MimeTypeTar,
	".gz":   MimeTypeGzip,
	".rar":  MimeTypeRar,
	".7z":   MimeType7z,
	".wasm": MimeTypeWasm,

	// Image extensions
	".jpg":  MimeTypeJPEG,
	".jpeg": MimeTypeJPEG,
	".png":  MimeTypePNG,
	".gif":  MimeTypeGIF,
	".webp": MimeTypeWebP,
	".svg":  MimeTypeSVG,
	".ico":  MimeTypeICO,
	".bmp":  MimeTypeBMP,
	".tif":  MimeTypeTIFF,
	".tiff": MimeTypeTIFF,

	// Audio extensions
	".mp3":  MimeTypeMP3,
	".wav":  MimeTypeWAV,
	".ogg":  MimeTypeOgg,
	".aac":  MimeTypeAAC,
	".flac": MimeTypeFlac,

	// Video extensions
	".mp4":  MimeTypeMP4,
	".webm": MimeTypeWebMVideo,
	".mpeg": MimeTypeMPEG,
	".avi":  MimeTypeAVI,
	".mov":  MimeTypeQuickTime,
	".mkv":  MimeTypeMatroska,
	".flv":  MimeTypeFlv,
}

// MimeTypeFromExtension returns MIME type for the given file extension.
// If the extension is not recognized, it returns MimeTypeOctetStream as a fallback.
func MimeTypeFromExtension(ext string) MimeType {
	if mimeType, ok := MimeTypeMapping[ext]; ok {
		return mimeType
	}

	return MimeTypeOctetStream
}
