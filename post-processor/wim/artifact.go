package wim

import (
	"fmt"
	"os"
)

// Internal implementation for packer.Artifact interface.
type WimArtifact struct {
	Name        string
	Path        string
	Compression uint32
}

func (w *WimArtifact) BuilderId() string {
	return "packer.post-processor.wim"
}

func (w *WimArtifact) Id() string {
	return ""
}

func (w *WimArtifact) Files() []string {
	return []string{w.Path}
}

func (w *WimArtifact) String() string {
	return fmt.Sprintf("New .wim file create in: %s", w.Path)
}

func (w *WimArtifact) State(name string) interface{} {
	return nil
}

func (w *WimArtifact) Destroy() error {
	os.Remove(w.Path)
	return nil
}
