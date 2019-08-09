// +build dev

package clients

import (
	ioutils "github.com/bbernhard/imagemonkey-core/ioutils"
	"fmt"
)

func NewLabelsDownloader(repository string, downloadLocation string) *LabelsDownloader {
	return &LabelsDownloader{
		repositoryUrl: repository,
		downloadLocation: downloadLocation,
	}
}

func (p *LabelsDownloader) Download() error {
	return ioutils.CopyDirectory(p.repositoryUrl, p.downloadLocation)
}
