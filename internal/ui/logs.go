package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"strings"
)

type Logs struct {
	Time    string
	Level   string
	Message string
	Source  string
}

func BuildLogsItem(log Logs) fyne.CanvasObject {
	time := widget.NewLabel(log.Time)
	level := widget.NewLabel(log.Level)
	message := widget.NewLabel(log.Message)
	source := widget.NewLabel(log.Source)

	return container.NewBorder(
		// top, bottom, left, right, center
		nil,
		nil,
		time,
		source,
		container.NewHBox(level, message),
	)
}

func BuildLogs() fyne.CanvasObject {
	title := widget.NewLabel("Logs")
	title.TextStyle = fyne.TextStyle{Bold: true}

	// for the test
	logList := []Logs{
		{
			Time:    "18:32:01",
			Level:   "INFO",
			Message: "Application started",
			Source:  "Application",
		},
		{
			Time:    "18:32:04",
			Level:   "WARN",
			Message: "Server connection not tested",
			Source:  "Connection",
		},
		{
			Time:    "11:32:04",
			Level:   "WARN",
			Message: "test",
			Source:  "test",
		},
	}

	listContainer := container.NewVBox()

	refreshLogList := func(logs []Logs) {
		listContainer.RemoveAll()

		for _, log := range logs {
			item := BuildLogsItem(log)
			listContainer.Add(item)
			listContainer.Add(widget.NewSeparator())
		}

		listContainer.Refresh()
	}

	for _, log := range logList {
		item := BuildLogsItem(log)
		listContainer.Add(item)
		listContainer.Add(widget.NewSeparator())
	}

	// Currently Selected Level
	selectedLevel := "ALL"

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Search Logs...")

	// Link the search term filter and the level filter
	applyLogFilters := func() {
		matched := FilterLogsLevel(logList, selectedLevel)
		matched = FilterLogs(matched, searchEntry.Text)
		refreshLogList(matched)
	}

	searchEntry.OnChanged = func(s string) {
		applyLogFilters()
	}

	allButton := widget.NewButton(
		"All",
		func() {
			selectedLevel = "ALL"
			applyLogFilters()
		},
	)

	infoButton := widget.NewButton(
		"Info",
		func() {
			selectedLevel = "INFO"
			applyLogFilters()
		},
	)
	infoButton.Importance = widget.SuccessImportance

	warnButton := widget.NewButton(
		"Warning",
		func() {
			selectedLevel = "WARN"
			applyLogFilters()
		},
	)
	warnButton.Importance = widget.WarningImportance

	errorButton := widget.NewButton(
		"Error",
		func() {
			selectedLevel = "ERROR"
			applyLogFilters()
		},
	)
	errorButton.Importance = widget.DangerImportance

	clearButton := widget.NewButton(
		"Clear Logs",
		func() {
			// to do : clear logs
		},
	)

	exportButton := widget.NewButton(
		"Export Logs",
		func() {
			// to do : export logs
		},
	)

	filterButtons := container.NewGridWithColumns(
		4,
		allButton,
		infoButton,
		warnButton,
		errorButton,
	)

	topContent := container.NewVBox(
		title,
		widget.NewSeparator(),
		searchEntry,
		filterButtons,
		widget.NewSeparator(),
	)

	bottomContent := container.NewVBox(
		clearButton,
		exportButton,
	)

	scrollableLogs := container.NewVScroll(listContainer)

	logs := container.NewBorder(
		topContent,
		bottomContent,
		nil,
		nil,
		scrollableLogs,
	)

	return logs
}

func FilterLogs(logs []Logs, keyword string) []Logs {
	result := []Logs{}

	keyword = strings.ToLower(keyword)

	for _, log := range logs {
		time := strings.ToLower(log.Time)
		message := strings.ToLower(log.Message)
		source := strings.ToLower(log.Source)

		if strings.Contains(time, keyword) ||
			strings.Contains(message, keyword) ||
			strings.Contains(source, keyword) {
			result = append(result, log)
		}
	}

	return result
}

func FilterLogsLevel(logs []Logs, level string) []Logs {
	result := []Logs{}

	for _, log := range logs {
		if level == "ALL" || strings.EqualFold(log.Level, level) {
			result = append(result, log)
		}
	}

	return result
}
