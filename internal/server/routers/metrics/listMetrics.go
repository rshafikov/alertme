package metrics

import (
	"github.com/rshafikov/alertme/internal/server/errmsg"
	"github.com/rshafikov/alertme/internal/server/logger"
	"github.com/rshafikov/alertme/internal/server/models"
	"html/template"
	"net/http"
)

// ListMetrics generates an HTML page displaying a table of all available metrics.
// It queries the store for metrics, converts them to plain format, and renders them using a template.
// The response is sent as an HTML document with HTTP status 200 on success.
// In case of errors, it logs and sends an appropriate HTTP error response.
func (h *Router) ListMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metrics := h.store.List(ctx)

	var plainMetrics []models.PlainMetric
	for _, metric := range metrics {
		plainMetrics = append(plainMetrics, *metric.ConvertToPlain())
	}

	const tmpl = `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Metrics</title>
		<style>
			body { font-family: Arial, sans-serif; margin: 20px; }
			table { border-collapse: collapse; width: 50%; }
			th, td { border: 1px solid black; padding: 8px; text-align: left; }
			th { background-color: #f2f2f2; }
		</style>
	</head>
	<body>
		<h1>Metrics List</h1>
		<table>
			<tr><th>Type</th><th>Name</th><th>Value</th></tr>
			{{range .}}
				<tr><td>{{.Type}}</td><td>{{.Name}}</td><td>{{.Value}}</td></tr>
			{{end}}
		</table>
	</body>
	</html>`

	t, err := template.New("metrics").Parse(tmpl)
	if err != nil {
		logger.Log.Debug(errmsg.UnableToParseTemplate)
		http.Error(w, errmsg.UnableToParseTemplate, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	err = t.Execute(w, plainMetrics)
	if err != nil {
		logger.Log.Debug(errmsg.UnableToWriteTemplate)
		http.Error(w, errmsg.UnableToWriteTemplate, http.StatusInternalServerError)
	}
}
