package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"text/template"

	owm "github.com/briandowns/openweathermap"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

var apiKey = os.Getenv("OWM_API_KEY")
var locale = os.Getenv("WEATHER_LOCALE")
var countryCode = os.Getenv("WEATHER_COUNTRY_CODE")

var locationZIP int64

var grid = ui.NewGrid()
var headLine = widgets.NewParagraph()
var leftBox = widgets.NewParagraph()
var rightBox = widgets.NewParagraph()

var fmap = template.FuncMap{
	"formatAsTimeFromInt":      formatAsTimeFromInt,
	"formatAsTimeFromTime":     formatAsTimeFromTime,
	"formatAsDateTimeFromTime": formatAsDateTimeFromTime,
}

const weatherTemplate = `


   Now: {{.Main.Temp}} °C
 Feels: {{.Main.FeelsLike}} °C
-----------------
  High: {{.Main.TempMax}} °C
   Low: {{.Main.TempMin}} °C
-----------------
   Hum: {{.Main.Humidity}} %
 Press: {{.Main.Pressure}} hPa
-----------------
  rise: {{.Sys.Sunrise | formatAsTimeFromInt}}
   set: {{.Sys.Sunset | formatAsTimeFromInt}}
`

const headLineTemplate = `{{.Name}} - {{range .Weather}}{{.Main}} ({{.Description}}) {{end}}`

const forecastTemplate = `{{$first := true}}{{range .List}}{{if not $first}}--------------------------------{{else}}{{$first = false}}{{end}}
 {{.DtTxt.Time | formatAsTimeFromTime}} - {{range .Weather}}{{.Main}} ({{.Description}}){{end}}
 T:{{.Main.Temp}} F:{{.Main.FeelsLike}} H:{{.Main.TempMax}} L:{{.Main.TempMin}}
{{end}}
`

func main() {
	locationZIP, _ = strconv.ParseInt(os.Getenv("WEATHER_LOCATION_ZIP"), 10, 64)

	if locationZIP == 0 {
		log.Fatal("WEATHER_LOCATION_ZIP not set")
		os.Exit(1)
	}

	if locale == "" {
		log.Fatal("WEATHER_LOCALE not set")
		os.Exit(1)
	}

	if countryCode == "" {
		log.Fatal("WEATHER_COUNTRY_CODE not set")
		os.Exit(1)
	}

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
		os.Exit(1)
	}
	defer ui.Close()
	setupTui()

	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(60 * time.Second).C
	updateTUI()

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				ui.Close()
				os.Exit(0)
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				grid.SetRect(0, 0, payload.Width, payload.Height)
				ui.Clear()
				ui.Render(grid)
			}
		case <-ticker:
			updateTUI()
		}
	}
}

func setupTui() {
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)
	grid.Set(
		ui.NewRow(1.0/20, headLine),
		ui.NewRow(19.0/20,
			ui.NewCol(4.0/11, leftBox),
			ui.NewCol(7.0/11, rightBox),
		),
	)

	headLine.Border = false
	headLine.TitleStyle = ui.NewStyle(ui.ColorGreen, ui.ColorBlack, ui.ModifierBold)

	leftBox.TitleStyle = headLine.TitleStyle
	leftBox.BorderStyle = ui.NewStyle(ui.ColorGreen, ui.ColorBlack)
	leftBox.TextStyle = ui.NewStyle(ui.ColorGreen, ui.ColorBlack)
	leftBox.WrapText = false
	leftBox.Title = " weather "

	rightBox.TitleStyle = headLine.TitleStyle
	rightBox.BorderStyle = leftBox.BorderStyle
	rightBox.TextStyle = leftBox.TextStyle
	rightBox.WrapText = leftBox.WrapText
	rightBox.Title = " forecast "
}

func updateTUI() {
	updateCurrent()
	updateForecast()

	ui.Render(grid)
}

func updateForecast() {
	var b bytes.Buffer
	f, err := owm.NewForecast("5", "C", locale, apiKey)
	if err != nil {
		log.Fatalln(err)
	}
	err = f.DailyByZip(int(locationZIP), countryCode, 5)
	if err != nil {
		log.Fatalln(err)
	}

	forecast := f.ForecastWeatherJson.(*owm.Forecast5WeatherData)

	tmpl, err := template.New("weather").Funcs(fmap).Parse(forecastTemplate)
	if err != nil {
		log.Fatalln(err)
	}

	if err := tmpl.Execute(&b, forecast); err != nil {
		log.Fatalln(err)
	}

	rightBox.Text = b.String()
}

func updateCurrent() {
	var b bytes.Buffer
	w, err := owm.NewCurrent("C", locale, apiKey)
	if err != nil {
		log.Fatalln(err)
	}
	err = w.CurrentByZip(int(locationZIP), countryCode)
	if err != nil {
		log.Fatalln(err)
	}

	tmpl, err := template.New("weather").Funcs(fmap).Parse(weatherTemplate)
	if err != nil {
		log.Fatalln(err)
	}

	if err := tmpl.Execute(&b, w); err != nil {
		log.Fatalln(err)
	}

	leftBox.Text = b.String()

	b.Reset()
	tmpl, err = template.New("weather").Funcs(fmap).Parse(headLineTemplate)
	if err != nil {
		log.Fatalln(err)
	}

	if err = tmpl.Execute(&b, w); err != nil {
		log.Fatalln(err)
	}

	headLine.Title = fmt.Sprintf(" %s ", b.String())
}

func formatAsTimeFromInt(t int) string {
	hour, min, _ := time.Unix(int64(t), 0).Clock()
	return fmt.Sprintf("%0.2d:%0.2d", hour, min)
}

func formatAsTimeFromTime(t time.Time) string {
	hour, min, _ := t.Clock()
	return fmt.Sprintf("%0.2d:%0.2d", hour, min)
}

func formatAsDateTimeFromTime(t time.Time) string {
	year, month, day := t.Date()
	hour, min, _ := t.Clock()
	return fmt.Sprintf("%0.4d/%0.2d/%0.2d %0.2d:%0.2d", year, month, day, hour, min)
}
