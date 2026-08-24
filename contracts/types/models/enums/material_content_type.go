package enums

type MaterialContentType string

const (
	MaterialContentType_None      MaterialContentType = "none"
	MaterialContentType_JSON      MaterialContentType = "application/json"
	MaterialContentType_PDF       MaterialContentType = "application/pdf"
	MaterialContentType_PlainText MaterialContentType = "text/plain"
	MaterialContentType_HTML      MaterialContentType = "text/html"
	MaterialContentType_Markdown  MaterialContentType = "text/markdown"
	MaterialContentType_PNG       MaterialContentType = "image/png"
	MaterialContentType_JPG       MaterialContentType = "image/jpg"
	MaterialContentType_JPEG      MaterialContentType = "image/jpeg"
	MaterialContentType_GIF       MaterialContentType = "image/gif"
	MaterialContentType_SVG       MaterialContentType = "image/svg+xml"
	MaterialContentType_WebP      MaterialContentType = "image/webp"
	MaterialContentType_MP4       MaterialContentType = "video/mp4"
	MaterialContentType_WebM      MaterialContentType = "video/webm"
	MaterialContentType_Mpeg      MaterialContentType = "audio/mpeg"
)
