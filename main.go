package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"
)

func genReport(config Config) {
	var (
		err error
		res = map[string]int{}
	)
	ctx := context.TODO()
	cli := GetGithubClient(ctx, config.GithubToken)

	startTime, err := GetStartTime(config)
	if err != nil {
		panic(err)
	}
	endTime, err := GetEndTime(config)
	if err != nil {
		panic(err)
	}

	var listOpts = &github.CommitsListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		Since:       startTime,
		Until:       endTime,
	}
	fmt.Println("[info] statistic since:", startTime, " until:", endTime)

	for _, repoName := range config.Repositories {
		res, err = counting(ctx, cli, repoName, listOpts, res)
		if err != nil {
			panic(err)
		}
	}

	commits := 0
	var contributors []string
	var newContributors []string
	for name, c := range res {
		commits = commits + c
		if name == "dependabot[bot]" {
			continue
		}
		contributors = append(contributors, name)
		total, _ := personTotal(ctx, cli, config, name)
		if c >= total {
			newContributors = append(newContributors, name)
		}
	}

	fmt.Println("contributors count: ", len(contributors), " commits: ", commits)
	fmt.Println("contributors: ", strings.Join(contributors, ","))
	fmt.Println("new contributors: ", strings.Join(newContributors, ","))
}

func GetStartTime(config Config) (time.Time, error) {
	if config.StartDate == "" {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0,
			0, 0, time.FixedZone("GMT", 8*3600)).AddDate(0, 0, -1*config.LastDays), nil
	}

	t, err := time.Parse("2006-01-02", config.StartDate)
	if err != nil {
		return time.Time{}, err
	}

	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0,
		0, 0, time.FixedZone("GMT", 8*3600))

	return t, err
}

func GetEndTime(config Config) (time.Time, error) {
	if config.EndDate == "" {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 23, 59,
			59, 0, time.FixedZone("GMT", 8*3600)).AddDate(0, 0, -1), nil
	}

	t, err := time.Parse("2006-01-02", config.EndDate)
	if err != nil {
		return time.Time{}, err
	}

	t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59,
		59, 0, time.FixedZone("GMT", 8*3600))

	return t, err
}

func SplitRepo(repo string) (string, string, error) {
	strs := strings.Split(repo, "/")
	if len(strs) != 2 {
		return "", "", fmt.Errorf("Repo format error")
	}
	return strs[0], strs[1], nil
}

func GetGithubClient(ctx context.Context, token string) *github.Client {
	tc := getToken(ctx, token)

	return github.NewClient(tc)
}

func getContributors(ctx context.Context, ghCli *github.Client,
	repoName string, listOpts *github.ListContributorsOptions, res map[string]string) (map[string]string, error) {
	owner, repo, err := SplitRepo(repoName)
	if err != nil {
		return nil, err
	}

	contributors, resp, err := ghCli.Repositories.ListContributors(ctx, owner, repo, listOpts)
	if err != nil {
		fmt.Printf("list contributors err: %s, %#v", err, resp)
	}

	for _, contributor := range contributors {
		if _, ok := res[*contributor.Login]; ok {
			continue
		}
		user, resp, err := ghCli.Users.Get(ctx, *contributor.Login)
		if err != nil {
			fmt.Printf("list contributors err: %s, %#v", err, resp)
			return nil, err
		}
		res[*contributor.Login] = ""
		if user.Email != nil {
			res[*contributor.Login] = *user.Email
		}
	}

	if len(contributors) >= 100 {
		page := listOpts.ListOptions.Page + 1
		listOpts.ListOptions.Page = page
		res, err = getContributors(ctx, ghCli, repoName, listOpts, res)
	} else {
		listOpts.ListOptions.Page = 1
	}

	return res, nil
}

func counting(ctx context.Context, ghCli *github.Client, repoName string, listOpts *github.CommitsListOptions, res map[string]int) (map[string]int, error) {
	owner, repo, err := SplitRepo(repoName)
	if err != nil {
		return res, err
	}

	commits, resp, err := ghCli.Repositories.ListCommits(ctx, owner, repo, listOpts)
	if err != nil {
		fmt.Printf("list commits err: %s, %#v", err, resp)
	}

	for _, commit := range commits {
		var login string
		if *commit.Author != nil {
			fmt.Println("commit:", commit)
			login := *commit.Author.Login
		}
		if _, ok := res[login]; ok {
			res[login] = res[login] + 1
		} else {
			res[login] = 1
		}
	}

	return res, nil
}

func personTotal(ctx context.Context, ghCli *github.Client, config Config, login string) (int, string) {
	count := 0
	var email string
	for _, repoName := range config.Repositories {
		tc, e, err := countingPersonal(ctx, ghCli, repoName, login)
		if email == "" && e != "" && !strings.Contains(e, "@users.noreply.github.com") {
			email = e
		}
		if err != nil {
			fmt.Println("countingPersonal err:", err, " user: ", login, " repo: ", repoName)
		}
		count = count + tc
	}

	return count, email
}

func countingPersonal(ctx context.Context, ghCli *github.Client, repoName, login string) (int, string, error) {
	owner, repo, err := SplitRepo(repoName)
	if err != nil {
		return 0, "", err
	}
	var listOpts = &github.CommitsListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		Author:      login,
	}

	commits, resp, err := ghCli.Repositories.ListCommits(ctx, owner, repo, listOpts)
	if err != nil {
		fmt.Printf("list commits err: %s, %#v", err, resp)
		return 0, "", err
	}
	var email string
	if len(commits) > 0 {
		email = *commits[0].Commit.Author.Email
	}

	return len(commits), email, nil
}

func getToken(ctx context.Context, token string) *http.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return tc
}
