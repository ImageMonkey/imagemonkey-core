// +build !dev

package clients

import (
	"os"
	"github.com/go-git/go-git/v5"
)

func NewLabelsDownloader(repositoryUrl string, downloadLocation string) *LabelsDownloader {
	return &LabelsDownloader{
		repositoryUrl: repositoryUrl,
		downloadLocation: downloadLocation,
	}
}

func (p *LabelsDownloader) Download() error {
	os.RemoveAll(p.downloadLocation)
	_, err := git.PlainClone(p.downloadLocation, false, &git.CloneOptions{
		URL:      p.repositoryUrl,
		Progress: os.Stdout,
	})
	return err
}
