// +build dev

package commons

type internalState int

const (
    Created = iota
    Started
   	Passed 
    
)

type TravisCiApi struct {
	token string
	state internalState
}


func NewTravisCiApi(repoOwner string, repo string) *TravisCiApi {
	return &TravisCiApi{
		state: Created,
	}
}

func (p *TravisCiApi) SetToken(token string) {
	p.token = token
}

func (p *TravisCiApi) GetBuildInfo(branchName string) (CiBuildInfo, error) {
	var ciBuildInfo CiBuildInfo

	if p.state == Created {
		p.state = Started
		ciBuildInfo.LastBuild.State = "created"
	} else if p.state == Started {
		p.state = Passed
		ciBuildInfo.LastBuild.State = "started"
	} else if p.state == Passed {
		ciBuildInfo.LastBuild.State = "passed"
	}

	return ciBuildInfo, nil
}

func (p *TravisCiApi) StartBuild(branchName string) error {
	return nil
}

