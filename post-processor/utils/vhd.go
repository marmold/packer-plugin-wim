package utils

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func MountImageVHD(context context.Context, sourceImage string, mountDir string) error {
	err := exec.CommandContext(context, "cmd", "/c", "dism", "/mount-image", strings.Join([]string{"/imagefile", sourceImage}, ":"), "/Index:1", strings.Join([]string{"/mountdir", mountDir}, ":")).Run()
	if err != nil {
		return fmt.Errorf("Unable to mount image %s to mount dir: %s", sourceImage, mountDir)
	}
	return nil
}

// We do not provide here context as this will throw unspecific error when for plugin actions will be for example, canceled. Instead of that we forward the error to final return in post-processor.
func UnmountImageVHD(mountDir string) error {
	err := exec.Command("cmd", "/c", "dism", "/Unmount-image", strings.Join([]string{"/mountdir", mountDir}, ":"), "/Discard").Run()
	if err != nil {
		return fmt.Errorf("Failed to unmount directory: %s. Manual action needed to unmount directory", mountDir)
	}
	return nil
}
