package wrap

import (
	"archive/tar"
	"errors"
	"io"
	"sync"

	"github.com/google/go-containerregistry/pkg/gcrane"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/stream"
	"github.com/yolocs/wraptor/pkg/file"
)

type Wrapper struct {
	baseImage  string
	filePrefix string

	m           sync.Mutex
	layers      []v1.Layer
	cacheOnce   sync.Once
	cachedImage v1.Image
	cacheErr    error
}

type Option func(*Wrapper)

func WithBaseImage(baseImage string) Option {
	return func(w *Wrapper) {
		w.baseImage = baseImage
	}
}

func WithFilePrefix(filePrefix string) Option {
	return func(w *Wrapper) {
		w.filePrefix = filePrefix
	}
}

func NewWrapper(opts ...Option) *Wrapper {
	w := &Wrapper{}
	for _, o := range opts {
		o(w)
	}
	return w
}

var ErrImageCached = errors.New("cannot append files: image has already been cached")

func (w *Wrapper) AppendFiles(readers ...*file.Reader) error {
	if w.cachedImage != nil {
		return ErrImageCached
	}

	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		tw := tar.NewWriter(pw)
		defer tw.Close()

		for _, reader := range readers {
			hdr := &tar.Header{
				Name: w.filePrefix + reader.Name,
				Mode: 0644,
				Size: reader.Size,
			}
			if err := tw.WriteHeader(hdr); err != nil {
				pw.CloseWithError(err)
				return
			}
			if _, err := io.Copy(tw, reader); err != nil {
				pw.CloseWithError(err)
				return
			}
		}
	}()

	layer := stream.NewLayer(pr)

	w.m.Lock()
	defer w.m.Unlock()
	w.layers = append(w.layers, layer)
	return nil
}

func (w *Wrapper) RawImage() (v1.Image, error) {
	w.cacheOnce.Do(func() {
		var base v1.Image
		if w.baseImage == "" {
			base = empty.Image
		} else {
			baseRef, err := name.ParseReference(w.baseImage)
			if err != nil {
				w.cacheErr = err
				return
			}
			remoteBase, err := remote.Image(baseRef, remote.WithAuthFromKeychain(gcrane.Keychain))
			if err != nil {
				w.cacheErr = err
				return
			}
			base = remoteBase
		}

		w.cachedImage, w.cacheErr = mutate.AppendLayers(base, w.layers...)
	})

	return w.cachedImage, w.cacheErr
}

func (w *Wrapper) ToRemote(image string) error {
	img, err := w.RawImage()
	if err != nil {
		return err
	}

	targetRef, err := name.ParseReference(image)
	if err != nil {
		return err
	}

	return remote.Write(targetRef, img, remote.WithAuthFromKeychain(gcrane.Keychain))
}

func (w *Wrapper) ToOCIArchive(path string) error {
	w.m.Lock()
	layerCount := len(w.layers)
	w.m.Unlock()

	if layerCount == 0 {
		return nil
	}

	img, err := w.RawImage()
	if err != nil {
		return err
	}

	layoutPath, err := layout.Write(path, empty.Index)
	if err != nil {
		return err
	}

	return layoutPath.AppendImage(img)
}

func (w *Wrapper) ToDaemon(t string) error {
	img, err := w.RawImage()
	if err != nil {
		return err
	}

	tag, err := name.NewTag(t)
	if err != nil {
		return err
	}

	_, err = daemon.Write(tag, img)
	return err
}
