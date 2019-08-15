// +build dev

package clients

func NewLabelsDownloader(repository string, downloadLocation string) *LabelsDownloader {
	return &LabelsDownloader{
		repositoryUrl: repository,
		downloadLocation: downloadLocation,
	}
}

func (p *LabelsDownloader) Download() error {
	return nil
}
