package handler

import (
	"html/template"
	"net/http"
)

var indexTmpl = template.Must(template.New("index").Parse(`<!doctype html>
<html>
<head><meta charset="utf-8"><title>Metrics</title></head>
<body>
  <h1>Metrics</h1>
  <h2>Gauges</h2>
  <ul>
    {{range $name, $value := .Gauges}}
      <li>{{$name}}: {{$value}}</li>
    {{else}}
      <li>No gauges</li>
    {{end}}
  </ul>
  <h2>Counters</h2>
  <ul>
    {{range $name, $value := .Counters}}
      <li>{{$name}}: {{$value}}</li>
    {{else}}
      <li>No counters</li>
    {{end}}
  </ul>
</body>
</html>`))

type metricsPageData struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

func renderMetricsHTML(w http.ResponseWriter, gauges map[string]float64, counters map[string]int64) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = indexTmpl.Execute(w, metricsPageData{Gauges: gauges, Counters: counters})
}
