package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"go.massi.dev/webcamtimelapse/internal/runner"
)

func showCalculatorDialog(w fyne.Window, intervalEntry, fpsEntry, framesEntry *widget.Entry) {
	interval, err := strconv.ParseFloat(intervalEntry.Text, 64)
	if err != nil || interval <= 0 {
		interval = 120
	}
	fps, err := strconv.ParseFloat(fpsEntry.Text, 64)
	if err != nil || fps <= 0 {
		fps = 30
	}
	frames, err := strconv.ParseFloat(framesEntry.Text, 64)
	if err != nil || frames < 0 {
		frames = 0
	}

	calcIntervalEntry := widget.NewEntry()
	calcIntervalEntry.SetText(fmt.Sprintf("%v", interval))

	calcFpsEntry := widget.NewEntry()
	calcFpsEntry.SetText(fmt.Sprintf("%v", fps))

	calcFramesEntry := widget.NewEntry()
	calcFramesEntry.SetText(fmt.Sprintf("%v", frames))

	captureEntry := widget.NewEntry()
	captureEntry.SetPlaceHolder("e.g., 24h, 2h30m")

	videoEntry := widget.NewEntry()
	videoEntry.SetPlaceHolder("e.g., 10s, 1m")

	var isUpdating bool

	updateFromN := func(n, i, f float64) {
		tc := time.Duration(n * i * float64(time.Second)).Round(time.Second)
		tv := time.Duration((n / f) * float64(time.Second)).Round(time.Millisecond)
		captureEntry.SetText(tc.String())
		videoEntry.SetText(tv.String())
	}

	getVals := func() (i, f, n float64) {
		i, _ = strconv.ParseFloat(calcIntervalEntry.Text, 64)
		if i <= 0 {
			i = 120
		}
		f, _ = strconv.ParseFloat(calcFpsEntry.Text, 64)
		if f <= 0 {
			f = 30
		}
		n, _ = strconv.ParseFloat(calcFramesEntry.Text, 64)
		if n < 0 {
			n = 0
		}
		return i, f, n
	}

	calcIntervalEntry.OnChanged = func(s string) {
		if isUpdating {
			return
		}
		isUpdating = true
		defer func() { isUpdating = false }()
		i, f, n := getVals()
		updateFromN(n, i, f)
	}

	calcFpsEntry.OnChanged = func(s string) {
		if isUpdating {
			return
		}
		isUpdating = true
		defer func() { isUpdating = false }()
		i, f, n := getVals()
		updateFromN(n, i, f)
	}

	calcFramesEntry.OnChanged = func(s string) {
		if isUpdating {
			return
		}
		isUpdating = true
		defer func() { isUpdating = false }()
		i, f, n := getVals()
		updateFromN(n, i, f)
	}

	captureEntry.OnChanged = func(s string) {
		if isUpdating {
			return
		}
		dur, err := time.ParseDuration(s)
		if err != nil {
			return
		}

		isUpdating = true
		defer func() { isUpdating = false }()

		i, f, _ := getVals()
		n := dur.Seconds() / i
		calcFramesEntry.SetText(fmt.Sprintf("%d", int(n)))

		tv := time.Duration((n / f) * float64(time.Second)).Round(time.Millisecond)
		videoEntry.SetText(tv.String())
	}

	videoEntry.OnChanged = func(s string) {
		if isUpdating {
			return
		}
		dur, err := time.ParseDuration(s)
		if err != nil {
			return
		}

		isUpdating = true
		defer func() { isUpdating = false }()

		i, f, _ := getVals()
		n := dur.Seconds() * f
		calcFramesEntry.SetText(fmt.Sprintf("%d", int(n)))

		tc := time.Duration(n * i * float64(time.Second)).Round(time.Second)
		captureEntry.SetText(tc.String())
	}

	isUpdating = true
	updateFromN(frames, interval, fps)
	isUpdating = false

	form := widget.NewForm(
		widget.NewFormItem("Capture Frequency (s)", calcIntervalEntry),
		widget.NewFormItem("Output Video FPS", calcFpsEntry),
		widget.NewFormItem("Number of Frames", calcFramesEntry),
		widget.NewFormItem("Capture Duration", captureEntry),
		widget.NewFormItem("Video Duration", videoEntry),
	)

	instructions := widget.NewLabel("Modify any value. Durations: 1h30m, 10s.")
	instructions.TextStyle.Italic = true

	content := container.NewVBox(
		instructions,
		form,
	)

	d := dialog.NewCustomConfirm("Duration Calculator", "Apply", "Cancel", content, func(apply bool) {
		if apply {
			intervalEntry.SetText(calcIntervalEntry.Text)
			fpsEntry.SetText(calcFpsEntry.Text)
			framesEntry.SetText(calcFramesEntry.Text)
		}
	}, w)
	d.Resize(fyne.NewSize(450, 350))

	d.Show()
}

func main() {
	app.SetMetadata(fyne.AppMetadata{
		ID:   "dev.massi.webcamtimelapse",
		Name: "WebCamTimeLapse",
	})
	a := app.NewWithID("dev.massi.webcamtimelapse")
	w := a.NewWindow("WebCamTimeLapse")
	w.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("WebCamTimeLapse"),
	))
	w.Resize(fyne.NewSize(600, 400))

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("http://...")

	intervalEntry := widget.NewEntry()
	intervalEntry.SetText("120")

	framesEntry := widget.NewEntry()
	framesEntry.SetText("0") // 0 = unlimited

	fpsEntry := widget.NewEntry()
	fpsEntry.SetText("30")

	calcBtn := widget.NewButton("Calc", func() {
		showCalculatorDialog(w, intervalEntry, fpsEntry, framesEntry)
	})

	inputForm := widget.NewForm(
		widget.NewFormItem("URL", urlEntry),
		widget.NewFormItem("Interval (s)", intervalEntry),
		widget.NewFormItem("Frames (0=inf)", container.NewBorder(nil, nil, nil, calcBtn, framesEntry)),
	)

	defaultOutput := runner.DefaultOutputFilename()
	fileEntry := widget.NewEntry()
	fileEntry.SetPlaceHolder(defaultOutput)

	browseBtn := widget.NewButton("Browse", func() {
		fd := dialog.NewFileSave(func(uc fyne.URIWriteCloser, err error) {
			if uc != nil {
				fileEntry.SetText(uc.URI().Path())
				_ = uc.Close()
			}
		}, w)
		fd.SetFileName(defaultOutput)
		fd.Show()
	})

	qualitySlider := widget.NewSlider(0.0, 1.0)
	qualitySlider.SetValue(1.0)
	qualitySlider.Step = 0.1
	qualityLabel := widget.NewLabel("1.0")
	qualitySlider.OnChanged = func(f float64) {
		qualityLabel.SetText(fmt.Sprintf("%.1f", f))
	}

	outputForm := widget.NewForm(
		widget.NewFormItem("File", container.NewBorder(nil, nil, nil, browseBtn, fileEntry)),
		widget.NewFormItem("FPS", fpsEntry),
		widget.NewFormItem("Quality", container.NewBorder(nil, nil, nil, qualityLabel, qualitySlider)),
	)

	statusLabel := widget.NewLabel("Ready")
	progressBar := widget.NewProgressBar()
	progressBar.Hide()
	infiniteBar := widget.NewProgressBarInfinite()
	infiniteBar.Hide()

	// isRunning is written from both the UI and capture goroutines.
	var isRunning atomic.Bool
	// cancelCapture is only read/written from the UI goroutine (OnTapped).
	var cancelCapture context.CancelFunc

	startBtn := widget.NewButton("Start", nil)
	startBtn.OnTapped = func() {
		if isRunning.Load() {
			if cancelCapture != nil {
				cancelCapture()
				startBtn.Disable()
				statusLabel.SetText("Finishing capture...")
			}
			return
		}

		url := urlEntry.Text
		if url == "" {
			dialog.ShowError(fmt.Errorf("URL cannot be empty"), w)
			return
		}

		interval, _ := strconv.Atoi(intervalEntry.Text)
		if interval <= 0 {
			interval = 120
		}
		frames, _ := strconv.Atoi(framesEntry.Text)
		fps, _ := strconv.Atoi(fpsEntry.Text)
		if fps <= 0 {
			fps = 30
		}
		file := fileEntry.Text
		if file == "" {
			file = runner.DefaultOutputFilename()
			if abs, err := filepath.Abs(file); err == nil {
				file = abs
			}
			fileEntry.SetText(file)
		}

		cfg := runner.Config{
			URL:      url,
			OutFile:  file,
			Interval: interval,
			Frames:   frames,
			FPS:      fps,
			Quality:  qualitySlider.Value,
		}

		isRunning.Store(true)
		startBtn.SetText("Stop")
		startBtn.Enable()
		if frames > 0 {
			progressBar.Show()
			progressBar.SetValue(0)
			infiniteBar.Hide()
			infiniteBar.Stop()
		} else {
			progressBar.Hide()
			infiniteBar.Show()
			infiniteBar.Start()
		}

		setupBar := widget.NewProgressBar()
		progressDialog := dialog.NewCustom("Setting up FFmpeg", "Cancel", setupBar, w)
		var dialogShown bool

		ctx, cancel := context.WithCancel(context.Background())
		cancelCapture = cancel

		progressFn := func(ev runner.ProgressEvent) {
			switch ev.Kind {
			case runner.EventSetup:
				fyne.Do(func() {
					if !dialogShown && ev.Pct < 100 {
						progressDialog.Show()
						dialogShown = true
					}
					setupBar.SetValue(float64(ev.Pct) / 100.0)
					if ev.Pct == 100 && dialogShown {
						progressDialog.Hide()
						statusLabel.SetText("Starting capture")
					}
				})
			case runner.EventCapture:
				fyne.Do(func() {
					dur := time.Duration(float64(ev.FrameCount) / float64(fps) * float64(time.Second)).Round(100 * time.Millisecond)
					statusLabel.SetText(fmt.Sprintf("Captured frame %d (expected video duration: %s)", ev.FrameCount, dur))
					if frames > 0 {
						progressBar.SetValue(float64(ev.FrameCount) / float64(frames))
					}
				})
			case runner.EventCompile:
				fyne.Do(func() {
					statusLabel.SetText("Compiling video...")
					startBtn.Disable()
					progressBar.Hide()
					infiniteBar.Show()
					infiniteBar.Start()
				})
			}
		}

		go func() {
			err := runner.RunCapture(ctx, cfg, progressFn)
			cancel() // release context resources
			isRunning.Store(false)
			fyne.Do(func() {
				startBtn.Enable()
				startBtn.SetText("Start")
				progressBar.Hide()
				infiniteBar.Stop()
				infiniteBar.Hide()
				if dialogShown {
					progressDialog.Hide()
				}
				if err != nil {
					dialog.ShowError(err, w)
					statusLabel.SetText("Error occurred")
				} else {
					statusLabel.SetText("Done!")
				}
			})
		}()
	}

	w.SetContent(container.NewVBox(
		widget.NewLabelWithStyle("Input", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		inputForm,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Output", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		outputForm,
		layout.NewSpacer(),
		widget.NewSeparator(),
		container.NewBorder(nil, nil, startBtn, nil, container.NewVBox(statusLabel, progressBar, infiniteBar)),
	))
	w.ShowAndRun()
}
