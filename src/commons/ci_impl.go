// +build !dev

package commons

import (
	"errors"
	"gopkg.in/resty.v1"
	"strconv"
	"time"
)

type GithubActionsApi struct {
	repoOwner string
	repo      string
	token     string
}

type Workflow struct {
	Id int64 `json:"id"`
}

type WorkflowRun struct {
	Id         int64  `json:"id"`
	HtmlUrl    string `json:"html_url"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HeadBranch string `json:"head_branch"`
	Event      string `json:"event"`
}

type WorkflowsResult struct {
	TotalCount int        `json:"total_count"`
	Workflows  []Workflow `json:"workflows"`
}

type WorkflowRunsResult struct {
	TotalCount   int           `json:"total_count"`
	WorkflowRuns []WorkflowRun `json:"workflow_runs"`
}

func NewGithubActionsApi(repoOwner string, repo string) *GithubActionsApi {
	return &GithubActionsApi{
		repoOwner: repoOwner,
		repo:      repo,
	}
}

func (p *GithubActionsApi) SetToken(token string) {
	p.token = token
}

func (p *GithubActionsApi) getBuilds(branchName string) (WorkflowRunsResult, error) {
	var res WorkflowRunsResult

	url := "https://api.github.com/repos/" + p.repoOwner + "/" + p.repo + "/actions/runs"

	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeader("Authorization", "Bearer "+p.token).
		SetPathParams(map[string]string{
			"branch": branchName,
		}).
		SetResult(&res).
		Get(url)
	if err != nil {
		return res, err
	}

	if !((resp.StatusCode() >= 200) && (resp.StatusCode() <= 209)) {
		return res, errors.New(resp.String())
	}
	return res,nil
}

func (p *GithubActionsApi) GetBuildInfo(branchName string) (CiBuildInfo, error) {
	var githubActionsBuildInfo CiBuildInfo

	builds, err := p.getBuilds(branchName)
	if err != nil {
		return githubActionsBuildInfo, err
	}

	if builds.TotalCount == 0 {
		return githubActionsBuildInfo, errors.New("No workflow run found")
	}

	githubActionsBuildInfo.JobUrl = builds.WorkflowRuns[0].HtmlUrl

	status := builds.WorkflowRuns[0].Status
	conclusion := builds.WorkflowRuns[0].Conclusion

	if status == "queued" {
		githubActionsBuildInfo.LastBuild.State = "created"
	} else if status == "in_progress" {
		githubActionsBuildInfo.LastBuild.State = "started"
	} else if status == "completed" && conclusion == "failure" {
		githubActionsBuildInfo.LastBuild.State = "failed"
	} else if status == "completed" && conclusion == "success" {
		githubActionsBuildInfo.LastBuild.State = "passed"
	} else if status == "completed" && conclusion == "cancelled" {
		githubActionsBuildInfo.LastBuild.State = "canceled"
	} else {
		githubActionsBuildInfo.LastBuild.State = "failed"
	}
	githubActionsBuildInfo.LastBuild.Id = builds.WorkflowRuns[0].Id
	return githubActionsBuildInfo, nil
}

func (p *GithubActionsApi) getWorkflowId() (string, error) {
	url := "https://api.github.com/repos/" + p.repoOwner + "/" + p.repo + "/actions/workflows"

	var res WorkflowsResult

	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeader("Authorization", "Bearer "+p.token).
		SetResult(&res).
		Get(url)
	if err != nil {
		return "", err
	}

	if !((resp.StatusCode() >= 200) && (resp.StatusCode() <= 209)) {
		return "", errors.New(resp.String())
	}

	if res.TotalCount != 1 {
		return "", errors.New("Couldn't determine workflow - got more than one workflow")
	}

	return strconv.FormatInt(res.Workflows[0].Id, 10), nil
}

func (p *GithubActionsApi) StartBuild(branchName string) error {
	type GithubActionsRequest struct {
		Ref string `json:"ref"`
	}

	workflowId, err := p.getWorkflowId()
	if err != nil {
		return err
	}

	buildsBefore, err := p.getBuilds(branchName)
	if err != nil {
		return err
	}

	numBuildsBefore := buildsBefore.TotalCount

	var req GithubActionsRequest
	req.Ref = branchName

	url := "https://api.github.com/repos/" + p.repoOwner + "/" + p.repo + "/actions/workflows/" + workflowId + "/dispatches"

	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeader("Authorization", "Bearer "+p.token).
		SetBody(&req).
		Post(url)

	if err != nil {
		return err
	}

	if !((resp.StatusCode() >= 200) && (resp.StatusCode() <= 209)) {
		return errors.New(resp.String())
	}

	//there seems to be no way (yet) to get the build number of the job we manually start.
	//so in order to determine the build number, we look at the build that was created at last. 
	//that's not bulletproof, but it should work well enough.
	for i := 0; i < 5; i++ { //it might take a bit until the build number is available
		buildsAfter, err := p.getBuilds(branchName)
		if err != nil {
			time.Sleep(1*time.Second)
			continue
		}
		numBuildsAfter := buildsAfter.TotalCount
		if numBuildsAfter == numBuildsBefore + 1 {
			if buildsAfter.WorkflowRuns[0].HeadBranch == branchName && buildsAfter.WorkflowRuns[0].Event == "workflow_dispatch" {
				return nil
			}
		}

		time.Sleep(1*time.Second)
	}

	return errors.New("Couldn't get build info after 5 attempts")
}
