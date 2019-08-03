package commons

type CiBuildInfo struct {
	LastBuild struct {
		State string `json:"state"`
		Id    int64  `json:"id"`
	} `json:"last_build"`
	JobUrl string `json:"job_url"`
}

type CiApi interface {
	SetToken(token string)
	GetBuildInfo(branchName string) (CiBuildInfo, error) 
	StartBuild(branchName string) error 
}

