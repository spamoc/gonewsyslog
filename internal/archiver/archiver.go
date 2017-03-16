package archiver

import (
	"compress/gzip"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

type Archive func(from string, to string) error

func XZArchive(absfrom string, to string) error {
	//temporary implements
	base := filepath.Dir(to)
	from := filepath.Base(absfrom)
	err := exec.Command("tar", "cJf", to, "-C", base, from).Run()
	if err != nil {
		return err
	}
	err = os.Remove(absfrom)
	return err
}

func BZ2Archive(absfrom string, to string) error {
	//temporary implements
	base := filepath.Dir(to)
	from := filepath.Base(absfrom)
	err := exec.Command("tar", "cjf", to, "-C", base, from).Run()
	if err != nil {
		return err
	}
	err = os.Remove(absfrom)
	return err
}

func GZArchive(from string, to string) error {
	stats, _ := os.Stat(from)
	mode := stats.Mode()
	tof, err := os.OpenFile(to, os.O_WRONLY|os.O_CREATE|os.O_APPEND, mode)
	if err != nil {
		return err
	}
	defer tof.Close()
	writer, err := gzip.NewWriterLevel(tof, gzip.BestCompression)
	if err != nil {
		return err
	}
	defer writer.Close()

	b, err := ioutil.ReadFile(from)
	if err != nil {
		return err
	}
	_, err = writer.Write(b)
	if err != nil {
		return err
	}
	if err := os.Remove(from); err != nil {
		return err
	}
	return nil
}
