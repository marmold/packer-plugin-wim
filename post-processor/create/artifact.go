package wimcreate

import (
	"fmt"
	"os"
)

type WimArtifact struct {
	Name string
	Path string
}

func (w *WimArtifact) BuilderId() string {
	return "packer.post-processor.wimcreate"
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
