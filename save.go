package main

import (
	"context"
	"fmt"
	"image/png"
	"os"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/stewi1014/glfractal/programs"
)

type SaveOptions struct {
	Name          string
	Width, Height int
	Antialias     float32
	Multithread   bool
}

func save(
	ctx context.Context,
	window gtk.IWindow,
	opts SaveOptions,
	program programs.Program,
	uniforms programs.Uniforms,
) {
	var err error
	ctx, cancel := WithErrorDialogCancelCause(window, ctx)
	defer CatchPanicToContext(cancel)

	file, err := os.Create(opts.Name)
	if err != nil {
		cancel(err)
		return
	}
	context.AfterFunc(ctx, func() {
		file.Close()
	})
	keepFile := context.AfterFunc(ctx, func() {
		os.Remove(file.Name())
	})

	var progressWindow ProgressBarWindow
	if opts.Multithread {
		app, err := window.ToWindow().GetApplication()
		if err != nil {
			cancel(err)
			return
		}

		progressWindow, err = NewImagePreviewWindow(
			ctx, app, "Save Image",
			func() { keepFile() },
			func() { cancel(context.Canceled) },
		)
		if err != nil {
			cancel(err)
			return
		}
	} else {
		progressWindow, err = NewProgressBarDialog(
			ctx, window, "Save Image",
			fmt.Sprintf("Saving %v", file.Name()),
			func() { cancel(context.Canceled) },
		)
		if err != nil {
			cancel(err)
			return
		}
	}

	image, err := program.GetImage(uniforms, opts.Width, opts.Height)
	if err != nil {
		cancel(err)
		return
	}

	go func() {
		defer CatchPanicToContext(cancel)
		if opts.Antialias > 0 {
			image = AntiAlias9x(image, opts.Antialias)
		}

		imageImage := ToImage(image)

		if opts.Multithread {
			progressWindow.AddProgressSupplier(ctx, WrapWithProgress(&imageImage), "Rendering to Buffer")
			buff := BufferImage(imageImage)
			imageImage = buff
			progressWindow.AddProgressSupplier(ctx, WrapWithProgress(&imageImage), "Encoding PNG")

			if imageWindow, ok := progressWindow.(*ImagePreviewWindow); ok {
				imageWindow.SetImageSupplier(ctx, func(dest *gdk.Pixbuf) {
					buff.Scale(dest, dest.GetWidth(), dest.GetHeight(), gdk.INTERP_NEAREST)
				})
			}

			err = buff.Buffer(ctx)
			if err != nil {
				cancel(err)
				return
			}
		} else {
			progressWindow.AddProgressSupplier(ctx, WrapWithProgress(&imageImage), "Rendering to PNG")
		}

		err = png.Encode(file, imageImage)
		if err != nil {
			cancel(err)
			return
		}

		if progressBarDialog, ok := progressWindow.(*ProgressBarDialog); ok {
			glib.IdleAdd(func() {
				progressBarDialog.Destroy()

				app, err := window.ToWindow().GetApplication()
				if err != nil {
					cancel(err)
					return
				}

				previewWindow, err := NewImagePreviewWindow(
					ctx, app, "Save Image",
					func() { keepFile() },
					func() { cancel(context.Canceled) },
				)
				if err != nil {
					cancel(err)
					return
				}

				previewWindow.OpenImage(file.Name())
				previewWindow.SetFinished("Finished")
			})
		}
	}()
}
