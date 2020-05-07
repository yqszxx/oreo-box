package image

import (
	"fmt"
	"github.com/yqszxx/oreo-box/config"
	"io/ioutil"
	"os"
	"text/tabwriter"
)

func List() error {
	files, err := ioutil.ReadDir(config.ImagePath)
	if err != nil {
		return fmt.Errorf("cannot read dir %s: %v", config.ImagePath, err)
	}

	var images []string
	for _, file := range files {
		if file.IsDir() {
			images = append(images, file.Name())
		}
	}

	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	if _, err := fmt.Fprint(w, "NAME\n"); err != nil {
		return fmt.Errorf("fail to exec fmt.Fprint : %v", err)
	}
	for _, item := range images {
		_, err := fmt.Fprintf(w, "%s\n", item)
		if err != nil {
			return fmt.Errorf("fail to exec fmt.Fprintf %v", err)
		}
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("cannot flush : %v", err)
	}
	return nil
}
