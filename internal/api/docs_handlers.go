package api

import (
	"net/http"
	um "uptime-monitor"
)

func (app *Application) GetYamlSpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/yaml")
	w.Write(um.OpenAPISpec)
}

func (app *Application) ShowSwaggerUI(w http.ResponseWriter, r *http.Request) {
	html := `
	<!doctype html>
	<html>
	    <head>
			<title>API Reference</title>
			<meta charset="utf-8"/>
			<meta
			    name="viewport"
				content="width=device-width, initial-scale=1"/>
		</head>
		<body>
		    <div id="app"></div>
			<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
			<script>
			    Scalar.createApiReference('#app',{
    				url:'/openapi.yaml'
				})
			</script>
		</body>
	</html>
	`
	w.Header().Add("Content-Type", "text/html")
	w.Write([]byte(html))
}
