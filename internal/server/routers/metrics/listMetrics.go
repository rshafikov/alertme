package metrics

import (
	"html/template"
	"log"
	"net/http"
)

func (h *Router) ListMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.store.List()
	if err != nil {
		log.Println("ERR", "cannot list metrics in storage", err)
		http.Error(w, "cannot list metrics in storage", http.StatusInternalServerError)
		return
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
			<tr><th>Name</th><th>Value</th></tr>
			{{range .}}
				<tr><td>{{.Name}}</td><td>{{.Value}}</td></tr>
			{{end}}
		</table>
	</body>
	</html>`

	t, err := template.New("metrics").Parse(tmpl)
	if err != nil {
		log.Println("ERR", "cannot parse template", err)
		http.Error(w, "cannot parse template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	err = t.Execute(w, metrics)
	if err != nil {
		log.Println("ERR", "cannot execute template", err)
		http.Error(w, "cannot execute template", http.StatusInternalServerError)
	}
}
