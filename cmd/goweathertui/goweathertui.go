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
var topBox = widgets.NewParagraph()
var bottomBox = widgets.NewParagraph()

var fmap = template.FuncMap{
	"formatAsTime":     formatAsTime,
	"formatAsDateTime": formatAsDateTime,
}

const weatherTemplate = `

  Cond:{{range .Weather}} {{.Description}} {{end}}
   Now: {{.Main.Temp}} °C
 Feels: {{.Main.FeelsLike}} °C
-------------------------
  High: {{.Main.TempMax}} °C
   Low: {{.Main.TempMin}} °C
-------------------------
   Hum: {{.Main.Humidity}} %
 Press: {{.Main.Pressure}} hPa
-------------------------
  rise: {{.Sys.Sunrise | formatAsTime}}
   set: {{.Sys.Sunset | formatAsTime}}
`
const forecastTemplate = `
{{range .List}}D: {{.DtTxt.Time | formatAsDateTime}}
C: {{range .Weather}}{{.Main}} {{.Description}}{{end}}
T: {{.Main.Temp}} H: {{.Main.TempMax}} L: {{.Main.TempMin}}
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
			ui.NewCol(1.0/2, topBox),
			ui.NewCol(1.0/2, bottomBox),
		),
	)

	headLine.Border = false
	headLine.TitleStyle = ui.NewStyle(ui.ColorGreen, ui.ColorBlack, ui.ModifierBold)

	topBox.TitleStyle = headLine.TitleStyle
	topBox.BorderStyle = ui.NewStyle(ui.ColorGreen, ui.ColorBlack)
	topBox.TextStyle = ui.NewStyle(ui.ColorGreen, ui.ColorBlack)
	topBox.WrapText = false

	bottomBox.TitleStyle = headLine.TitleStyle
	bottomBox.BorderStyle = topBox.BorderStyle
	bottomBox.TextStyle = topBox.TextStyle
	bottomBox.WrapText = topBox.WrapText
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
	f.DailyByZip(int(locationZIP), countryCode, 5)
	forecast := f.ForecastWeatherJson.(*owm.Forecast5WeatherData)

	tmpl, err := template.New("weather").Funcs(fmap).Parse(forecastTemplate)
	if err != nil {
		log.Fatalln(err)
	}

	if err := tmpl.Execute(&b, forecast); err != nil {
		log.Fatalln(err)
	}

	bottomBox.Title = " forecast "
	bottomBox.Text = b.String()
}

func updateCurrent() {
	var b bytes.Buffer
	w, err := owm.NewCurrent("C", locale, apiKey)
	if err != nil {
		log.Fatalln(err)
	}
	w.CurrentByZip(int(locationZIP), countryCode)

	tmpl, err := template.New("weather").Funcs(fmap).Parse(weatherTemplate)
	if err != nil {
		log.Fatalln(err)
	}

	if err := tmpl.Execute(&b, w); err != nil {
		log.Fatalln(err)
	}

	headLine.Title = fmt.Sprintf(" %s ", w.Name)

	topBox.Title = " weather "
	topBox.Text = b.String()
}

func formatAsTime(t int) string {
	hour, min, _ := time.Unix(int64(t), 0).Clock()
	return fmt.Sprintf("%0.2d:%0.2d", hour, min)
}

func formatAsDateTime(t time.Time) string {
	year, month, day := t.Date()
	hour, min, _ := t.Clock()
	return fmt.Sprintf("%0.4d/%0.2d/%0.2d %0.2d:%0.2d", year, month, day, hour, min)
}
