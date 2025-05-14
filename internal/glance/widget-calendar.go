package glance

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"time"
)

var calendarWidgetTemplate = mustParseTemplate("calendar.html", "widget-base.html")

var calendarWeekdaysToInt = map[string]time.Weekday{
	"sunday":    time.Sunday,
	"monday":    time.Monday,
	"tuesday":   time.Tuesday,
	"wednesday": time.Wednesday,
	"thursday":  time.Thursday,
	"friday":    time.Friday,
	"saturday":  time.Saturday,
}

type calendarWidget struct {
	widgetBase     `yaml:",inline"`
	FirstDayOfWeek string        `yaml:"first-day-of-week"`
	FirstDay       int           `yaml:"-"`
	cachedHTML     template.HTML `yaml:"-"`
}

// GetID implements widget.
// Subtle: this method shadows the method (widgetBase).GetID of calendarWidget.widgetBase.
func (widget *calendarWidget) GetID() uint64 {
	panic("unimplemented")
}

// GetName implements widget.
// Subtle: this method shadows the method (widgetBase).GetName of calendarWidget.widgetBase.
func (widget *calendarWidget) GetName() string {
	panic("unimplemented")
}

// GetSecret implements widget.
func (widget *calendarWidget) GetSecret(name string) (string, error) {
	panic("unimplemented")
}

// GetType implements widget.
// Subtle: this method shadows the method (widgetBase).GetType of calendarWidget.widgetBase.
func (widget *calendarWidget) GetType() string {
	panic("unimplemented")
}

// handleRequest implements widget.
// Subtle: this method shadows the method (widgetBase).handleRequest of calendarWidget.widgetBase.
func (widget *calendarWidget) handleRequest(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// requiresUpdate implements widget.
// Subtle: this method shadows the method (widgetBase).requiresUpdate of calendarWidget.widgetBase.
func (widget *calendarWidget) requiresUpdate(*time.Time) bool {
	panic("unimplemented")
}

// setHideHeader implements widget.
// Subtle: this method shadows the method (widgetBase).setHideHeader of calendarWidget.widgetBase.
func (widget *calendarWidget) setHideHeader(bool) {
	panic("unimplemented")
}

// setID implements widget.
// Subtle: this method shadows the method (widgetBase).setID of calendarWidget.widgetBase.
func (widget *calendarWidget) setID(uint64) {
	panic("unimplemented")
}

// setProviders implements widget.
// Subtle: this method shadows the method (widgetBase).setProviders of calendarWidget.widgetBase.
func (widget *calendarWidget) setProviders(*widgetProviders) {
	panic("unimplemented")
}

// update implements widget.
// Subtle: this method shadows the method (widgetBase).update of calendarWidget.widgetBase.
func (widget *calendarWidget) update(context.Context) {
	panic("unimplemented")
}

func (widget *calendarWidget) initialize() error {
	widget.withTitle("Calendar").withError(nil)

	if widget.FirstDayOfWeek == "" {
		widget.FirstDayOfWeek = "monday"
	} else if _, ok := calendarWeekdaysToInt[widget.FirstDayOfWeek]; !ok {
		return errors.New("invalid first day of week")
	}

	widget.FirstDay = int(calendarWeekdaysToInt[widget.FirstDayOfWeek])
	widget.cachedHTML = widget.renderTemplate(widget, calendarWidgetTemplate)

	return nil
}

func (widget *calendarWidget) Render() template.HTML {
	return widget.cachedHTML
}
