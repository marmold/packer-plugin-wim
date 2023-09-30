package wim

/*
#cgo LDFLAGS: -L../../lib/devel -llibwim -Wl,-rpath=../../lib/devel
#include "../../lib/devel/wimlib.h"
#include <stdlib.h>
*/
import "C"

/*
Important note. Usage of wimlib C types and functions require that correct DLL is installed on Windows machine in order to invoke them. A libwim-15.dll is shipped with Windows binaries and this must be included somewhere in order to run the plugin.
According to the https://knowledgebase.progress.com/articles/Article/P91669, when you open an executable file that requires some dynamically linked libraries not listed in the system registry, the Windows operating system will search for them in the following locations:

The directory of the executable file being called.
The current working directory from which the executable was called.
The %SystemRoot%\SYSTEM32 directory.
The %SystemRoot% directory.
The directories listed in the Path environment variable.
*/

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/hashicorp/packer-plugin-sdk/packer"
)

// As we want to have option to cancel this action if needed, we need to pass a context to function and run it as goroutine.
func CreateWimWindows(context context.Context, ui packer.Ui, mountDir string, newArtifact WimArtifact) error {

	errChannel := make(chan error)

	go func(chan error) {

		// Enumerate mount directory for sub-directories and create new array for those.
		var subDirs []string
		entries, err := os.ReadDir(mountDir)
		if err != nil {
			errChannel <- fmt.Errorf("Unable to enumerate entries in mounted path: %s. Error: %s", mountDir, err)
			return
		} else {
			for _, subDir := range entries {
				if subDir.IsDir() {
					subDirs = append(subDirs, subDir.Name())
				}
			}
		}

		// Create new C type array of C type wimlib_capture_source. It has to be static array of length enough to keep all the counted sub-directories.
		size := len(subDirs)                                                                                       // Declare a size of array
		arr := (*C.struct_wimlib_capture_source)(C.malloc(C.size_t(C.sizeof_struct_wimlib_capture_source * size))) // Declare a C object (array of wimlib_capture_source* struct)
		defer C.free(unsafe.Pointer(arr))                                                                          // Free memory of C type array at the end of program

		ps := unsafe.Slice(arr, size) // Declare a Go slice backed by C array in order to fill it with data.

		// Append structs in C array with correct data by referring using Golang slice.
		ui.Message(fmt.Sprintf("Adding sources to WIM struct"))
		for i, subDir := range subDirs {
			rootPath := filepath.Join(mountDir, subDir)
			ps[i].fs_source_path = (*C.ushort)(syscall.StringToUTF16Ptr(rootPath))
			ps[i].wim_target_path = (*C.ushort)(syscall.StringToUTF16Ptr("\\" + subDir))
			ps[i].reserved = (C.long)(0)
			ui.Message(fmt.Sprintf("\t+ '%s'", rootPath))
		}

		// Create WIM C struct
		var wim *C.WIMStruct
		val := C.wimlib_create_new_wim(newArtifact.Compression, &wim) // Use here only 1 or 2 as compression type as 3 wont be supported by DISM at the time of applying image.
		if val != 0 {
			errChannel <- fmt.Errorf("Unable to create wim struct using wimlib_create_new_wim C function. Error: %d", val)
			return
		}
		defer C.wimlib_free(wim)
		defer C.wimlib_global_cleanup()

		// Add sources to WIM struct
		if val := C.wimlib_add_image_multisource(
			wim,
			arr,
			C.size_t(size),
			(*C.ushort)(syscall.StringToUTF16Ptr(newArtifact.Name)),
			nil,
			67716,
		); val != 0 {
			ui.Message(fmt.Sprintf("Value: %d", val))
			errChannel <- fmt.Errorf("Unable to add to wim struct using wimlib_add_image_multisource C function. Error: %d", val)
			return
		}

		wimPath := filepath.Join(newArtifact.Path, newArtifact.Name+".wim")
		// wimPath := newArtifact.Path + "\\" + newArtifact.Name + ".wim"
		ui.Message(fmt.Sprintf("Successfully added sources to wim struct. Writing WIM to '%s'", wimPath))

		// Write the WIM to disk using corresponding WIM struct
		if val := C.wimlib_write(
			wim,
			(*C.ushort)(syscall.StringToUTF16Ptr(wimPath)),
			-1,
			1,
			0,
		); val != 0 {
			errChannel <- fmt.Errorf("Unable to write WIM using wimlib_write C function. Error: %d", val)
			return
		}

		// If everything is ok then return nil to channel.
		errChannel <- nil
	}(errChannel)

	select {
	case <-context.Done():
		return context.Err()
	case err := <-errChannel:
		return err
	}
}
