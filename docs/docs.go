package docs

import (
	"embed"
	"net/http"
)

//go:embed swagger.json swagger.yaml
var SwaggerFS embed.FS

// ServeSwaggerUI serves the Swagger UI from the embedded file system
func ServeSwaggerUI(mux *http.ServeMux, basePath string) {
	fileServer := http.FileServer(http.FS(SwaggerFS))
	mux.Handle(basePath+"/", http.StripPrefix(basePath, fileServer))
}