// +build dev
// +build cifailure

package commons

type GithubActionsApi struct {
	token string
}


func NewGithubActionsApi(repoOwner string, repo string) *GithubActionsApi {
	return &GithubActionsApi {
	}
}

func (p *GithubActionsApi) SetToken(token string) {
	p.token = token
}

func (p *GithubActionsApi) GetBuildInfo(branchName string) (CiBuildInfo, error) {
	var ciBuildInfo CiBuildInfo
	ciBuildInfo.LastBuild.State = "failed"

	return ciBuildInfo, nil
}

func (p *GithubActionsApi) StartBuild(branchName string) error {
	return nil
}

