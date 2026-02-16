package core

import (
	"dsoob/backend/tools"
	"fmt"
	"os"
	"path"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

func DebugImageResizer() {
	t := time.Now()

	root := path.Join(tools.DATA_DIRECTORY, "debug")
	list, err := os.ReadDir(root)
	if err != nil {
		fmt.Printf("Directory Error: %s\n", err.Error())
		return
	}

	var resizeLimit = make(chan int, max(runtime.NumCPU()-1, 1))
	var resizeCount atomic.Int64
	var resizeGroup sync.WaitGroup

	for i, entry := range list {
		if !entry.Type().IsRegular() {
			resizeGroup.Done()
			return
		}
		resizeGroup.Add(1)

		go func() {
			defer resizeGroup.Done()
			resizeLimit <- 1
			defer func() { <-resizeLimit }()

			filename := entry.Name()
			filepath := path.Join(root, filename)
			raw, err := os.ReadFile(filepath)
			if err != nil {
				fmt.Printf("Cannot Read '%s': %s\n", filepath, err.Error())
				return
			}

			// Resize Images
			t := time.Now()
			if _, err := tools.ImageProcessor(tools.ImageOptionsAvatars, int64(i), raw); err != nil {
				fmt.Printf("Cannot process '%s' as avatar: %s\n", filepath, err.Error())
				return
			}

			at := time.Since(t)
			t = time.Now()

			if _, err := tools.ImageProcessor(tools.ImageOptionsBanners, int64(i), raw); err != nil {
				fmt.Printf("Cannot process '%s' as banner: %s\n", filepath, err.Error())
				return
			}
			bt := time.Since(t)

			resizeCount.Add(1)
			fmt.Printf("%-96s (Avatar: %4dms) (Banner: %4dms)\n", filename, at.Milliseconds(), bt.Milliseconds())
		}()

	}

	resizeGroup.Wait()
	fmt.Printf("Completed %d items in %s\n", resizeCount.Load(), time.Since(t))

}
