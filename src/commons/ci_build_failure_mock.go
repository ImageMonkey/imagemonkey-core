// +build dev
// +build cifailure

package commons

type TravisCiApi struct {
	token string
}


func NewTravisCiApi(repoOwner string, repo string) *TravisCiApi {
	return &TravisCiApi {
	}
}

func (p *TravisCiApi) SetToken(token string) {
	p.token = token
}

func (p *TravisCiApi) GetBuildInfo(branchName string) (CiBuildInfo, error) {
	var ciBuildInfo CiBuildInfo
	ciBuildInfo.LastBuild.State = "failed"

	return ciBuildInfo, nil
}

func (p *TravisCiApi) StartBuild(branchName string) error {
	return nil
}

