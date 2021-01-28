// +build !dev

package commons

import (
	"gopkg.in/resty.v1"
	"errors"
	"strconv"
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
	Id int64 `json:"id"`
	HtmlUrl string `json:"html_url"`
	Status string `json:"status"`
	Conclusion string `json:"conclusion"`
}

type WorkflowsResult struct {
	TotalCount int `json:"total_count"`
	Workflows []Workflow `json:"workflows"`
}


type WorkflowRunsResult struct {
	TotalCount int `json:"total_count"`
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

func (p *GithubActionsApi) GetBuildInfo(branchName string) (CiBuildInfo, error) {
	var githubActionsBuildInfo CiBuildInfo

	url := "https://api.github.com/repos/" + p.repoOwner + "/" + p.repo + "/actions/runs"

	var res WorkflowRunsResult

	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeader("Authorization", "Bearer "+p.token).
		SetPathParams(map[string]string {
			"branch": branchName,
		}).
		SetResult(&res).
		Get(url)
	if err != nil {
		return githubActionsBuildInfo, err
	}

	if !((resp.StatusCode() >= 200) && (resp.StatusCode() <= 209)) {
		return githubActionsBuildInfo, errors.New(resp.String())
	}

	if res.TotalCount == 0 {
		return githubActionsBuildInfo, errors.New("No workflow run found")
	}

	githubActionsBuildInfo.JobUrl = res.WorkflowRuns[res.TotalCount-1].HtmlUrl

	status := res.WorkflowRuns[res.TotalCount-1].Status
	conclusion := res.WorkflowRuns[res.TotalCount-1].Conclusion

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
	githubActionsBuildInfo.LastBuild.Id = res.WorkflowRuns[res.TotalCount-1].Id
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

	return nil
}

