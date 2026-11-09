package ui

import (
	"io"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"vpn-manager/internal/log"
)

func BuildLogsItem(log log.Log) fyne.CanvasObject {
	time := widget.NewLabel(log.Time)
	level := widget.NewLabel(log.Level)
	message := widget.NewLabel(log.Message)
	message.Wrapping = fyne.TextWrapWord
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

func BuildLogs(w fyne.Window) fyne.CanvasObject {
	title := widget.NewLabel("Logs")
	title.TextStyle = fyne.TextStyle{Bold: true}

	logList, err := log.ReadLogs()
	if err != nil {
		logList = []log.Log{}
	}

	listContainer := container.NewVBox()

	refreshLogList := func(logs []log.Log) {
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

	refresh := func() {
		newLogs, err := log.ReadLogs()
		if err != nil {
			handleError(w, "Failed to read logs", err)
			return
		}

		logList = newLogs
		applyLogFilters()
	}

	clearButton := widget.NewButton(
		"Clear Logs",
		func() {
			if err := os.WriteFile(log.LogFile, nil, 0644); err != nil {
				handleError(w, "Failed to clear logs", err)
				return
			}

			refresh()
		},
	)

	refreshButton := widget.NewButton(
		"Refresh Logs",
		func() {
			refresh()
		},
	)

	exportButton := widget.NewButton(
		"Export Logs",
		func() {
			saveDialog := dialog.NewFileSave(
				func(writer fyne.URIWriteCloser, err error) {
					if err != nil {
						handleError(w, "Failed to export logs", err)
						return
					}

					// If the user cancels the save dialog, writer will be nil
					if writer == nil {
						return
					}
					defer writer.Close()

					// Copy the contents of the log file to the selected location
					src, err := os.Open(log.LogFile)
					if err != nil {
						handleError(w, "Failed to open logs", err)
						return
					}
					defer src.Close()

					// Copy the contents of the log file to the selected location
					if _, err := io.Copy(writer, src); err != nil {
						handleError(w, "Failed to export logs", err)
						return
					}

					dialog.ShowInformation(
						"Export",
						"Logs exported successfully.",
						w,
					)
				},
				w,
			)
			saveDialog.SetFileName(".txt")
			// Set the filter to only allow .txt and .log files
			saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".txt", ".log"}))
			saveDialog.Show()
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

	bottomContent := container.NewGridWithColumns(
		3,
		refreshButton,
		clearButton,
		exportButton,
	)

	scrollableLogs := container.NewVScroll(listContainer)

	logs := container.NewBorder(
		topContent,
		container.NewVBox(
			widget.NewSeparator(),
			bottomContent,
		),
		nil,
		nil,
		scrollableLogs,
	)

	return logs
}

func FilterLogs(logs []log.Log, keyword string) []log.Log {
	result := []log.Log{}

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

func FilterLogsLevel(logs []log.Log, level string) []log.Log {
	result := []log.Log{}

	for _, log := range logs {
		if level == "ALL" || strings.EqualFold(log.Level, level) {
			result = append(result, log)
		}
	}

	return result
}
