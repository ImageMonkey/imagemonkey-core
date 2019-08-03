// +build dev

package clients

func NewLabelsDownloader(repositoryUrl string, downloadLocation string) *LabelsDownloader {
	return &LabelsDownloader{
		repositoryUrl: repositoryUrl,
		downloadLocation: downloadLocation,
	}
}

func (p *LabelsDownloader) Download() error {	
	return nil
}
