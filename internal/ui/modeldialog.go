package ui

import (
	"context"
	"fmt"

	"github.com/diamondburned/gotk4-adwaita/pkg/adw"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"

	"github.com/storo/guanaco/internal/logger"
	"github.com/storo/guanaco/internal/ollama"
)

// ModelDialog is a dialog for downloading Ollama models.
type ModelDialog struct {
	*adw.Window

	// UI components
	entry        *gtk.Entry
	progressBar  *gtk.ProgressBar
	statusLabel  *gtk.Label
	downloadBtn  *gtk.Button
	cancelBtn    *gtk.Button
	modelListBox *gtk.ListBox

	// State
	client        *ollama.Client
	cancelFunc    context.CancelFunc
	isDownloading bool
	models        []ollama.RegistryModel

	// Callbacks
	onModelDownloaded func(string)
}

// NewModelDialog creates a new model download dialog.
func NewModelDialog(parent *gtk.Window, client *ollama.Client) *ModelDialog {
	d := &ModelDialog{
		client: client,
	}

	d.Window = adw.NewWindow()
	d.SetTitle("Download Model")
	d.SetModal(true)
	d.SetDefaultSize(450, 500)
	if parent != nil {
		d.SetTransientFor(parent)
	}

	d.setupUI()

	return d
}

func (d *ModelDialog) setupUI() {
	// Header bar with close button
	headerBar := adw.NewHeaderBar()
	headerBar.SetShowEndTitleButtons(true)
	headerBar.SetShowStartTitleButtons(true)
	headerBar.SetTitleWidget(gtk.NewLabel("Download Model"))

	// Main content box
	content := gtk.NewBox(gtk.OrientationVertical, 12)
	content.SetMarginTop(16)
	content.SetMarginBottom(24)
	content.SetMarginStart(24)
	content.SetMarginEnd(24)

	// Available models label
	availableLabel := gtk.NewLabel("Available Models:")
	availableLabel.SetXAlign(0)
	content.Append(availableLabel)

	// Model list box
	d.modelListBox = gtk.NewListBox()
	d.modelListBox.SetSelectionMode(gtk.SelectionSingle)
	d.modelListBox.AddCSSClass("boxed-list")
	d.modelListBox.ConnectRowActivated(func(row *gtk.ListBoxRow) {
		idx := row.Index()
		if idx >= 0 && idx < len(d.models) {
			d.entry.SetText(d.models[idx].Name)
		}
	})

	scrolled := gtk.NewScrolledWindow()
	scrolled.SetChild(d.modelListBox)
	scrolled.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scrolled.SetMinContentHeight(180)
	scrolled.SetMaxContentHeight(220)
	scrolled.SetVExpand(true)
	content.Append(scrolled)

	// Load models in background
	go d.loadAvailableModels()

	// Custom model label
	customLabel := gtk.NewLabel("Or enter custom model:")
	customLabel.SetXAlign(0)
	customLabel.SetMarginTop(8)
	customLabel.AddCSSClass("dim-label")
	content.Append(customLabel)

	// Model name entry
	d.entry = gtk.NewEntry()
	d.entry.SetPlaceholderText("Model name...")
	d.entry.ConnectActivate(func() {
		if !d.isDownloading {
			d.startDownload()
		}
	})
	content.Append(d.entry)

	// Progress bar (hidden initially)
	d.progressBar = gtk.NewProgressBar()
	d.progressBar.SetVisible(false)
	d.progressBar.SetShowText(true)
	content.Append(d.progressBar)

	// Status label (hidden initially)
	d.statusLabel = gtk.NewLabel("")
	d.statusLabel.SetVisible(false)
	d.statusLabel.AddCSSClass("dim-label")
	d.statusLabel.SetWrap(true)
	content.Append(d.statusLabel)

	// Button box
	buttonBox := gtk.NewBox(gtk.OrientationHorizontal, 8)
	buttonBox.SetHAlign(gtk.AlignEnd)
	buttonBox.SetMarginTop(12)

	// Cancel button
	d.cancelBtn = gtk.NewButton()
	d.cancelBtn.SetLabel("Cancel")
	d.cancelBtn.ConnectClicked(func() {
		if d.isDownloading && d.cancelFunc != nil {
			d.cancelFunc()
		} else {
			d.Close()
		}
	})
	buttonBox.Append(d.cancelBtn)

	// Download button
	d.downloadBtn = gtk.NewButton()
	d.downloadBtn.SetLabel("Download")
	d.downloadBtn.AddCSSClass("suggested-action")
	d.downloadBtn.ConnectClicked(d.startDownload)
	buttonBox.Append(d.downloadBtn)

	content.Append(buttonBox)

	// Use ToolbarView to add header bar
	toolbarView := adw.NewToolbarView()
	toolbarView.AddTopBar(headerBar)
	toolbarView.SetContent(content)

	d.SetContent(toolbarView)
}

func (d *ModelDialog) startDownload() {
	modelName := d.entry.Text()
	if modelName == "" {
		return
	}

	logger.Info("Starting model download", "model", modelName)

	// Setup UI for downloading
	d.isDownloading = true
	d.entry.SetSensitive(false)
	d.downloadBtn.SetSensitive(false)
	d.downloadBtn.SetLabel("Downloading...")
	d.progressBar.SetVisible(true)
	d.progressBar.SetFraction(0)
	d.statusLabel.SetVisible(true)
	d.statusLabel.SetText("Starting download...")

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	d.cancelFunc = cancel

	go func() {
		err := d.client.PullModel(ctx, modelName, func(status string, completed, total int64) {
			glib.IdleAdd(func() {
				if total > 0 {
					progress := float64(completed) / float64(total)
					d.progressBar.SetFraction(progress)
					d.progressBar.SetText(fmt.Sprintf("%.1f%%", progress*100))
				}
				d.statusLabel.SetText(status)
			})
		})

		glib.IdleAdd(func() {
			d.isDownloading = false
			d.cancelFunc = nil

			if err != nil {
				if err == context.Canceled {
					d.statusLabel.SetText("Download cancelled")
				} else {
					d.statusLabel.SetText(fmt.Sprintf("Error: %v", err))
					d.statusLabel.AddCSSClass("error")
				}
				d.resetUI()
				return
			}

			// Success
			logger.Info("Model downloaded successfully", "model", modelName)
			d.statusLabel.SetText("Download complete!")
			d.progressBar.SetFraction(1.0)
			d.progressBar.SetText("100%")

			if d.onModelDownloaded != nil {
				d.onModelDownloaded(modelName)
			}

			// Close dialog after short delay
			glib.TimeoutAdd(1000, func() bool {
				d.Close()
				return false
			})
		})
	}()
}

func (d *ModelDialog) resetUI() {
	d.entry.SetSensitive(true)
	d.downloadBtn.SetSensitive(true)
	d.downloadBtn.SetLabel("Download")
}

// OnModelDownloaded sets the callback for when a model is successfully downloaded.
func (d *ModelDialog) OnModelDownloaded(callback func(string)) {
	d.onModelDownloaded = callback
}

func (d *ModelDialog) loadAvailableModels() {
	models := ollama.FetchAvailableModels(context.Background())

	glib.IdleAdd(func() {
		d.models = models
		for _, model := range models {
			row := d.createModelRow(model.Name, model.Description)
			d.modelListBox.Append(row)
		}
	})
}

func (d *ModelDialog) createModelRow(name, desc string) *gtk.ListBoxRow {
	box := gtk.NewBox(gtk.OrientationHorizontal, 8)
	box.SetMarginTop(6)
	box.SetMarginBottom(6)
	box.SetMarginStart(12)
	box.SetMarginEnd(12)

	nameLabel := gtk.NewLabel(name)
	nameLabel.SetXAlign(0)
	nameLabel.AddCSSClass("heading")
	box.Append(nameLabel)

	spacer := gtk.NewBox(gtk.OrientationHorizontal, 0)
	spacer.SetHExpand(true)
	box.Append(spacer)

	descLabel := gtk.NewLabel(desc)
	descLabel.AddCSSClass("dim-label")
	descLabel.AddCSSClass("caption")
	box.Append(descLabel)

	row := gtk.NewListBoxRow()
	row.SetChild(box)
	return row
}
