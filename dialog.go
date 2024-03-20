package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

func CatchPanicToContext(ctxCancel context.CancelCauseFunc) {
	if v := recover(); v != nil {
		err, ok := v.(error)
		if !ok {
			err = fmt.Errorf("panic: %v", v)
		}
		err = fmt.Errorf("%w\n%v", err, string(debug.Stack()))
		if ctxCancel != nil {
			ctxCancel(err)
		}
	}
}

func WithErrorDialogCancelCause(window gtk.IWindow, ctx context.Context) (context.Context, context.CancelCauseFunc) {
	ctx, cancel := context.WithCancelCause(ctx)

	return ctx, func(err error) {
		cancel(err)
		err = context.Cause(ctx)
		if !errors.Is(err, context.Canceled) {
			log.Println(err)
			NewErrorDialog(window, err, 1)
		}
	}
}

func NewErrorDialog(
	parent gtk.IWindow,
	err error,
	traceSkip int,
) {
	_, file, line, ok := runtime.Caller(traceSkip + 1)

	fileLocation := "unknown file"
	if ok {
		fileLocation = fmt.Sprintf("%s:%v", file, line)
	}

	glib.IdleAdd(func() {
		dialog := gtk.MessageDialogNew(
			parent,
			gtk.DIALOG_DESTROY_WITH_PARENT,
			gtk.MESSAGE_ERROR,
			gtk.BUTTONS_CLOSE,
			"Error in %s: %s",
			fileLocation,
			err.Error(),
		)
		dialog.SetIcon(iconPixbuf)
		dialog.Connect("response", dialog.Destroy)

		messageArea, err := dialog.GetMessageArea()
		if err != nil {
			log.Println(err)

		} else {
			messageArea.GetChildren().Foreach(func(item interface{}) {
				if widget, ok := item.(*gtk.Widget); ok {
					l, err := gtk.WidgetToLabel(widget)
					if err != nil {
						return
					}

					l.SetSelectable(true)
				}
			})
		}

		dialog.SetKeepAbove(true)
		dialog.Run()
	})
}

type ProgressBarWindow interface {
	gtk.IWindow
	AddProgressSupplier(context.Context, func() float64, string)
}

func NewProgressBar(ctx context.Context) *ProgressBar {
	bar := &ProgressBar{}

	bar.progressBar, _ = gtk.ProgressBarNew()
	bar.progressBar.SetProperty("show-text", true)

	bar.label, _ = gtk.LabelNew("starting...")

	return bar
}

type progressSupplier struct {
	description string
	supplier    func() float64
}

type ProgressBar struct {
	progressBar    *gtk.ProgressBar
	label          *gtk.Label
	suppliers      []progressSupplier
	suppliersMutex sync.Mutex
}

func (dialog *ProgressBar) SetFinished(description string) {
	dialog.suppliersMutex.Lock()
	defer dialog.suppliersMutex.Unlock()
	dialog.suppliers = nil
	glib.IdleAdd(func() {
		dialog.progressBar.SetFraction(1)
		dialog.label.SetText(description)
	})
}

func (dialog *ProgressBar) AddProgressSupplier(ctx context.Context, supplier func() float64, description string) {
	dialog.suppliersMutex.Lock()
	defer dialog.suppliersMutex.Unlock()

	if dialog.suppliers == nil {
		go dialog.updateProgress(ctx)
	}

	dialog.suppliers = append(dialog.suppliers, progressSupplier{
		supplier:    supplier,
		description: description,
	})
}

func (dialog *ProgressBar) updateProgress(ctx context.Context) {
	ticker := time.NewTicker(time.Second / 30)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			dialog.suppliersMutex.Lock()
			if len(dialog.suppliers) == 0 {
				dialog.suppliersMutex.Unlock()
				continue
			}

			progress := float64(1)
			description := ""
			for i := 0; progress >= 1; i++ {
				if i >= len(dialog.suppliers) {
					description = "Finished"
					break
				}
				progress = dialog.suppliers[i].supplier()
				description = dialog.suppliers[i].description
			}
			dialog.suppliersMutex.Unlock()

			var wg sync.WaitGroup
			wg.Add(1)
			glib.IdleAdd(func() {
				dialog.progressBar.SetFraction(progress)
				dialog.label.SetText(description)
				wg.Done()
			})
			wg.Wait()
		case <-ctx.Done():
			return
		}
	}
}

func NewProgressBarDialog(
	ctx context.Context,
	window gtk.IWindow,
	title string,
	message string,
	onCancel func(),
) (*ProgressBarDialog, error) {
	var err error
	dialog := &ProgressBarDialog{}

	var buttons []interface{}
	if onCancel != nil {
		buttons = []interface{}{"Cancel", gtk.RESPONSE_CANCEL}
	}

	dialog.Dialog, err = gtk.DialogNewWithButtons(
		title,
		window,
		gtk.DIALOG_DESTROY_WITH_PARENT,
		buttons,
	)
	if err != nil {
		return nil, err
	}
	dialog.SetIcon(iconPixbuf)
	dialog.Connect("response", func(dialog *gtk.Dialog, response gtk.ResponseType) {
		if response == gtk.RESPONSE_CANCEL {
			onCancel()
			dialog.Destroy()
		}
	})

	dialog.ProgressBar = NewProgressBar(ctx)

	content, _ := dialog.GetContentArea()
	content.Add(dialog.label)
	content.Add(dialog.progressBar)

	dialog.ShowAll()
	context.AfterFunc(ctx, func() {
		glib.IdleAdd(func() {
			dialog.Destroy()
		})
	})
	return dialog, nil
}

type ProgressBarDialog struct {
	*gtk.Dialog
	*ProgressBar
}

func NewImagePreviewWindow(
	ctx context.Context,
	app *gtk.Application,
	title string,
	onSave func(),
	onDelete func(),
) (*ImagePreviewWindow, error) {
	var err error
	window := &ImagePreviewWindow{}

	window.ApplicationWindow, err = gtk.ApplicationWindowNew(app)
	if err != nil {
		return nil, err
	}
	window.SetIcon(iconPixbuf)
	window.SetTitle(title)
	destroySignal := window.Connect("destroy", onDelete)

	window.image, err = gtk.ImageNew()
	if err != nil {
		return nil, err
	}

	width, height := getDisplaySize()
	window.image.SetSizeRequest(int(float64(width)*.8), int(float64(height)*.8))

	saveButton, _ := gtk.ButtonNewWithLabel("Save")
	saveButton.Connect("clicked", func() {
		onSave()
		window.HandlerDisconnect(destroySignal)
		window.Destroy()
	})

	deleteButton, _ := gtk.ButtonNewWithLabel("Delete")
	deleteButton.Connect("clicked", func() {
		onDelete()
		window.HandlerDisconnect(destroySignal)
		window.Destroy()
	})

	window.ProgressBar = NewProgressBar(ctx)
	window.progressBar.SetVExpand(false)

	grid, _ := gtk.GridNew()
	grid.Attach(window.image, 0, 0, 10, 1)
	grid.Attach(saveButton, 0, 1, 1, 2)
	grid.Attach(window.label, 1, 1, 8, 1)
	grid.Attach(window.progressBar, 1, 2, 8, 1)
	grid.Attach(deleteButton, 9, 1, 1, 2)
	window.Add(grid)

	window.ShowAll()
	context.AfterFunc(ctx, func() {
		glib.IdleAdd(func() {
			window.Destroy()
		})
	})
	return window, nil
}

type ImagePreviewWindow struct {
	*gtk.ApplicationWindow
	*ProgressBar
	image *gtk.Image
}

func (window *ImagePreviewWindow) OpenImage(filename string) error {
	pixbuf, err := gdk.PixbufNewFromFileAtSize(
		filename,
		window.image.GetAllocatedWidth(),
		window.image.GetAllocatedHeight(),
	)
	if err != nil {
		return err
	}

	window.image.SetFromPixbuf(pixbuf)
	return nil
}

func (window *ImagePreviewWindow) SetImageSupplier(ctx context.Context, supplier func(dest *gdk.Pixbuf)) {
	pixbuf, err := gdk.PixbufNew(
		gdk.COLORSPACE_RGB,
		true,
		8,
		window.image.GetAllocatedWidth(),
		window.image.GetAllocatedHeight(),
	)

	if err != nil {
		log.Println(err)
	}

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		var wg sync.WaitGroup

		for {
			select {
			case <-ticker.C:
				supplier(pixbuf)
				wg.Add(1)
				glib.IdleAdd(func() {
					window.image.SetFromPixbuf(pixbuf)
					wg.Done()
				})
				wg.Wait()

			case <-ctx.Done():
				return
			}
		}
	}()
}
