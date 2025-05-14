package glance

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"time"
)

// NewsSource represents a single news source.
type NewsSource struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

// newsWidget implements the widget interface for news headlines.
type newsWidget struct {
	widgetBase `yaml:",inline"`
	Sources    []NewsSource   `yaml:"sources"`
	Headlines  []NewsHeadline `yaml:"headlines"`
}

// NewsHeadline represents a single news headline (placeholder for now).
type NewsHeadline struct {
	Title  string `yaml:"title"`
	URL    string `yaml:"url"`
	Source string `yaml:"source"`
}

// initialize loads any required config or sets up the widget.
func (w *newsWidget) initialize() error {
	// Debug: log sources loaded from YAML
	log.Printf("[newsWidget] initialize: loaded %d sources: %+v", len(w.Sources), w.Sources)
	return nil
}

// requiresUpdate determines if the widget needs to update.
func (w *newsWidget) requiresUpdate(now *time.Time) bool {
	return w.widgetBase.requiresUpdate(now)
}

// setProviders sets the widget providers.
func (w *newsWidget) setProviders(providers *widgetProviders) {
	w.widgetBase.setProviders(providers)
}

// setID sets the widget ID.
func (w *newsWidget) setID(id uint64) {
	w.widgetBase.setID(id)
}

// setHideHeader sets the HideHeader flag.
func (w *newsWidget) setHideHeader(value bool) {
	w.widgetBase.setHideHeader(value)
}

// GetType returns the widget type.
func (w *newsWidget) GetType() string {
	return "news"
}

// GetID returns the widget ID.
func (w *newsWidget) GetID() uint64 {
	return w.widgetBase.GetID()
}

// handleRequest handles HTTP requests for the widget.
func (w *newsWidget) handleRequest(resp http.ResponseWriter, req *http.Request) {
	http.Error(resp, "news widget does not support direct requests", http.StatusNotImplemented)
}

// update fetches news headlines from sources.
func (w *newsWidget) update(ctx context.Context) {
	// Debug: log sources before updating
	log.Printf("[newsWidget] update: sources: %+v", w.Sources)
	var headlines []NewsHeadline
	for _, src := range w.Sources {
		headlines = append(headlines, NewsHeadline{
			Title:  "Top stories from " + src.Name,
			URL:    src.URL,
			Source: src.Name,
		})
	}
	w.Headlines = headlines
	w.ContentAvailable = len(w.Headlines) > 0
	w.scheduleNextUpdate()
}

// Render renders the widget as HTML.
func (w *newsWidget) Render() template.HTML {
	const tpl = `
<div class="news-widget">
  {{- if .Title }}
    <h3>{{ .Title }}</h3>
  {{- end }}
  <ul>
    {{- range .Headlines }}
      <li>
        <a href="{{ .URL }}" target="_blank" rel="noopener">{{ .Title }}</a>
        <span style="color: #888; font-size: 0.9em;">({{ .Source }})</span>
      </li>
    {{- else }}
      <li>No news sources configured.</li>
    {{- end }}
  </ul>
</div>
`
	t := template.Must(template.New("newsWidget").Parse(tpl))
	return w.renderTemplate(w, t)
}
