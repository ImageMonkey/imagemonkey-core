package commons

import (
	"gopkg.in/resty.v1"
	"errors"
	"strconv"
)

type TravisCiBuildInfo struct {
	LastBuild struct {
		State string `json:"state"`
		Id    int64  `json:"id"`
	} `json:"last_build"`
	JobUrl string `json:"job_url"`
}

type TravisCiApi struct {
	repoOwner string
	repo      string
	token     string
}

func NewTravisCiApi(repoOwner string, repo string) *TravisCiApi {
	return &TravisCiApi{
		repoOwner: repoOwner,
		repo:      repo,
	}
}

func (p *TravisCiApi) SetToken(token string) {
	p.token = token
}

func (p *TravisCiApi) GetBuildInfo(branchName string) (TravisCiBuildInfo, error) {

	url := "https://api.travis-ci.org/repo/" + p.repoOwner + "%2F" + p.repo + "/branch/" + branchName

	var travisCiBuildInfo TravisCiBuildInfo

	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeader("Travis-API-Version", "3").
		SetHeader("Authorization", "token "+p.token).
		SetResult(&travisCiBuildInfo).
		Get(url)

	if err != nil {
		return travisCiBuildInfo, err
	}

	if !((resp.StatusCode() >= 200) && (resp.StatusCode() <= 209)) {
		return travisCiBuildInfo, errors.New(resp.String())
	}
	//log.Info(resp.String())
	travisCiBuildInfo.JobUrl = ("https://travis-ci.org/" + p.repoOwner + "/" + p.repo +
		"/builds/" + strconv.FormatInt(travisCiBuildInfo.LastBuild.Id, 10))
	return travisCiBuildInfo, nil
}

func (p *TravisCiApi) StartBuild(branchName string) error {
	type TravisRequest struct {
		Request struct {
			Branch string `json:"branch"`
		} `json:"request"`
	}

	var req TravisRequest
	req.Request.Branch = branchName

	url := "https://api.travis-ci.org/repo/" + p.repoOwner + "%2F" + p.repo + "/requests"

	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeader("Travis-API-Version", "3").
		SetHeader("Authorization", "token "+p.token).
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

