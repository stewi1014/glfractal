package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime"
	"runtime/debug"
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

func WrapErrorDialog(parent *gtk.ApplicationWindow, failable func() error) func() {
	return func() {
		err := failable()
		if err != nil {
			log.Println(err)
			glib.IdleAdd(func() {
				NewErrorDialog(parent, err)
			})
		}
	}
}

func AttachErrorDialog(parent *gtk.ApplicationWindow, ctx context.Context) {
	go func() {
		<-ctx.Done()
		err := context.Cause(ctx)
		if !errors.Is(err, context.Canceled) {
			log.Println(err)
			glib.IdleAdd(func() {
				NewErrorDialog(parent, err)
			})
		}
	}()
}

func NewErrorDialog(
	parent *gtk.ApplicationWindow,
	err error,
) {
	_, file, line, ok := runtime.Caller(1)

	fileLocation := "unknown file"
	if ok {
		fileLocation = fmt.Sprintf("%s:%v", file, line)
	}

	dialog := gtk.MessageDialogNew(
		parent,
		gtk.DIALOG_DESTROY_WITH_PARENT,
		gtk.MESSAGE_ERROR,
		gtk.BUTTONS_CLOSE,
		"Error in %s: %s",
		fileLocation,
		err.Error(),
	)

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
}

func NewProgressDialog(
	parentCtx context.Context,
	parentWindow gtk.IWindow,
	title string,
	description string,
	onCancel func(),
) (*ProgressDialog, error) {
	dialog := &ProgressDialog{}
	dialog.Dialog, _ = gtk.DialogNewWithButtons(
		title,
		parentWindow,
		gtk.DIALOG_DESTROY_WITH_PARENT,
		[]interface{}{"CANCEL", gtk.RESPONSE_CANCEL},
	)
	dialog.SetKeepAbove(true)
	dialog.Connect("response", func(dialog *gtk.Dialog, response gtk.ResponseType) {
		if response == gtk.RESPONSE_CANCEL {
			onCancel()
		}
	})

	ca, _ := dialog.GetContentArea()
	dialog.label, _ = gtk.LabelNew(description)
	ca.Add(dialog.label)

	dialog.progressBar, _ = gtk.ProgressBarNew()
	dialog.progressBar.SetProperty("show-text", true)
	dialog.progressBar.SetSizeRequest(500, 80)
	ca.Add(dialog.progressBar)

	go dialog.periodicUpdate(parentCtx)
	return dialog, nil
}

type ProgressDialog struct {
	*gtk.Dialog
	progressBar *gtk.ProgressBar
	label       *gtk.Label

	progressFuncs []func() float64
}

// AddProgressSupplier adds a supplier for progress information to the ProgressDialog.
// If more than one supplier is added, their values are averaged.
func (dialog *ProgressDialog) AddProgressSupplier(supplier func() float64) {
	glib.IdleAdd(func() {
		dialog.progressFuncs = append(dialog.progressFuncs, supplier)
	})
}

func (dialog *ProgressDialog) periodicUpdate(ctx context.Context) {
	ticker := time.NewTicker(time.Second / 10)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			glib.IdleAdd(func() {
				progress := float64(0)
				for _, progressFunc := range dialog.progressFuncs {
					progress += progressFunc()
				}
				progress = progress / float64(len(dialog.progressFuncs))
				dialog.progressBar.SetFraction(progress)
			})
		case <-ctx.Done():
			glib.IdleAdd(func() {
				dialog.Destroy()
			})
			return
		}
	}
}

func NewImageDialog(
	app *gtk.Application,
	pixbuf *gdk.Pixbuf,
	responseSave func(),
	responseDelete func(),
) (*ImagePreview, error) {
	w := &ImagePreview{}
	var err error

	w.ApplicationWindow, err = gtk.ApplicationWindowNew(app)
	if err != nil {
		return nil, err
	}

	previewImage, err := gtk.ImageNewFromPixbuf(pixbuf)
	if err != nil {
		return nil, err
	}

	previewImage.SetHExpand(true)
	previewImage.SetVExpand(true)

	deleteButton, _ := gtk.ButtonNewWithLabel("Delete")
	deleteButton.Connect("clicked", func(button *gtk.Button) {
		if responseDelete != nil {
			responseDelete()
		}
		w.Destroy()
	})

	saveButton, _ := gtk.ButtonNewWithLabel("Save")
	saveButton.Connect("clicked", func(button *gtk.Button) {
		if responseSave != nil {
			responseSave()
		}
		w.Destroy()
	})

	grid, _ := gtk.GridNew()
	grid.Attach(previewImage, 0, 0, 5, 1)
	grid.Attach(saveButton, 0, 1, 1, 1)
	grid.Attach(deleteButton, 4, 1, 1, 1)

	w.Add(grid)

	return w, nil
}

type ImagePreview struct {
	*gtk.ApplicationWindow
}
