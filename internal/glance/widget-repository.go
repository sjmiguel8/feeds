package glance

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type repositoryWidget struct {
	widgetBase            `yaml:",inline"`
	RequestedRepositories []string     `yaml:"repositories"`
	Token                 string       `yaml:"token"`
	PullRequestsLimit     int          `yaml:"pull-requests-limit"`
	IssuesLimit           int          `yaml:"issues-limit"`
	CommitsLimit          int          `yaml:"commits-limit"`
	Repositories          []repository `yaml:"-"`
}

// GetID implements widget.
// Subtle: this method shadows the method (widgetBase).GetID of repositoryWidget.widgetBase.
func (w *repositoryWidget) GetID() uint64 {
	return w.widgetBase.GetID()
}

// GetType implements widget.
// Subtle: this method shadows the method (widgetBase).GetType of repositoryWidget.widgetBase.
func (w *repositoryWidget) GetType() string {
	return "repository"
}

// Render implements widget.
func (w *repositoryWidget) Render() template.HTML {
	if len(w.Repositories) == 0 {
		return template.HTML("<div class=\"widget widget-repository\"><p>No repositories found</p></div>")
	}
	if len(w.Repositories) == 1 {
		return template.HTML(fmt.Sprintf("<div class=\"widget widget-repository\"><h2>%s</h2><p>Stars: %d</p><p>Forks: %d</p></div>",
			w.Repositories[0].Name, w.Repositories[0].Stars, w.Repositories[0].Forks))
	}
	if len(w.Repositories) > 1 {
		var b strings.Builder
		b.WriteString("<div class=\"widget widget-repository\"><h2>Repositories</h2><ul>")
		for _, repo := range w.Repositories {
			b.WriteString(fmt.Sprintf("<li><strong>%s</strong> - Stars: %d, Forks: %d</li>", repo.Name, repo.Stars, repo.Forks))
		}
		b.WriteString("</ul></div>")
		return template.HTML(b.String())
	}
	return template.HTML("<div class=\"widget widget-repository\"><p>No repositories found</p></div>")
}

// requiresUpdate implements widget.
// Subtle: this method shadows the method (widgetBase).requiresUpdate of repositoryWidget.widgetBase.
func (w *repositoryWidget) requiresUpdate(lastUpdate *time.Time) bool {
	return w.widgetBase.requiresUpdate(lastUpdate)
}

// setHideHeader implements widget.
// Subtle: this method shadows the method (widgetBase).setHideHeader of repositoryWidget.widgetBase.
func (w *repositoryWidget) setHideHeader(hide bool) {
	w.widgetBase.setHideHeader(hide)
}

// setID implements widget.
// Subtle: this method shadows the method (widgetBase).setID of repositoryWidget.widgetBase.
func (w *repositoryWidget) setID(id uint64) {
	w.widgetBase.setID(id)
}

// setProviders implements widget.
// Subtle: this method shadows the method (widgetBase).setProviders of repositoryWidget.widgetBase.
func (w *repositoryWidget) setProviders(providers *widgetProviders) {
	w.widgetBase.setProviders(providers)
}

// GetName returns the name of the widget, which is used in templates
func (w *repositoryWidget) GetName() string {
	return w.widgetBase.GetName()
}

// Custom UnmarshalYAML to ensure repositories field is mapped correctly
func (w *repositoryWidget) UnmarshalYAML(unmarshal func(any) error) error {
	type plain repositoryWidget
	aux := &struct {
		Repositories []string `yaml:"repositories"`
		*plain
	}{
		plain: (*plain)(w),
	}
	if err := unmarshal(aux); err != nil {
		return err
	}
	return nil
}

func (widget *repositoryWidget) initialize() error {
	log.Printf("Initializing repository widget with ID %d", widget.ID)
	if widget.Token == "" {
		widget.Token, _ = widget.Providers.GetSecret("GITHUB-TOKEN")
		if widget.Token == "" {
			return fmt.Errorf("no GitHub token provided")
		}
	}

	fmt.Printf("RequestedRepositories: %#v\n", widget.RequestedRepositories)
	widget.withTitle("Repository").withCacheDuration(1 * time.Hour)

	if widget.PullRequestsLimit == 0 || widget.PullRequestsLimit < -1 {
		widget.PullRequestsLimit = 3
	}

	if widget.IssuesLimit == 0 || widget.IssuesLimit < -1 {
		widget.IssuesLimit = 3
	}

	if widget.CommitsLimit == 0 || widget.CommitsLimit < -1 {
		widget.CommitsLimit = -1
	}

	// Filter out empty/whitespace-only repository names
	filtered := make([]string, 0, len(widget.RequestedRepositories))
	for _, repo := range widget.RequestedRepositories {
		repo = strings.TrimSpace(repo)
		if repo != "" && strings.Contains(repo, "/") {
			filtered = append(filtered, repo)
		}
	}
	widget.RequestedRepositories = filtered

	return nil
}

func (widget *repositoryWidget) update(ctx context.Context) {
	widget.Repositories = make([]repository, 0, len(widget.RequestedRepositories))
	if len(widget.RequestedRepositories) == 0 {
		widget.canContinueUpdateAfterHandlingErr(fmt.Errorf("no repositories requested"))
		return
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for _, repo := range widget.RequestedRepositories {
		// Repo validation already done in initialize, but double-check
		if repo == "" || !strings.Contains(repo, "/") {
			continue // skip invalid repository names
		}

		wg.Add(1)
		go func(repo string) {
			defer wg.Done()
			details, err := fetchRepositoryDetailsFromGithub(
				repo,
				string(widget.Token),
				widget.PullRequestsLimit,
				widget.IssuesLimit,
				widget.CommitsLimit,
			)
			mu.Lock()
			defer mu.Unlock()
			if err != nil && firstErr == nil {
				firstErr = err
			}
			widget.Repositories = append(widget.Repositories, details)
		}(repo)
	}
	wg.Wait()
	widget.canContinueUpdateAfterHandlingErr(firstErr)
}

func (widget *repositoryWidget) handleRequest(w http.ResponseWriter, r *http.Request) {
}

type repository struct {
	Name             string
	Stars            int
	Forks            int
	OpenPullRequests int
	PullRequests     []githubTicket
	OpenIssues       int
	Issues           []githubTicket
	LastCommits      int
	Commits          []githubCommitDetails
}

type githubTicket struct {
	Number    int
	CreatedAt time.Time
	Title     string
}

type githubCommitDetails struct {
	Sha       string
	Author    string
	CreatedAt time.Time
	Message   string
}

type githubRepositoryResponseJson struct {
	Name  string `json:"full_name"`
	Stars int    `json:"stargazers_count"`
	Forks int    `json:"forks_count"`
}

type githubTicketResponseJson struct {
	Count   int `json:"total_count"`
	Tickets []struct {
		Number    int    `json:"number"`
		CreatedAt string `json:"created_at"`
		Title     string `json:"title"`
	} `json:"items"`
}

type gitHubCommitResponseJson struct {
	Sha    string `json:"sha"`
	Commit struct {
		Author struct {
			Name string `json:"name"`
			Date string `json:"date"`
		} `json:"author"`
		Message string `json:"message"`
	} `json:"commit"`
}

func fetchRepositoryDetailsFromGithub(repo string, token string, maxPRs int, maxIssues int, maxCommits int) (repository, error) {
	// Validate repository name format
	if repo == "" || !strings.Contains(repo, "/") {
		return repository{
			Name: repo,
		}, fmt.Errorf("%w: invalid repository format: %s", errNoContent, repo)
	}

	repositoryRequest, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s", repo), nil)
	if repositoryRequest == nil {
		return repository{
			Name: repo,
		}, fmt.Errorf("%w: could not create request with repository: %v", errNoContent, err)
	}

	PRsRequest, _ := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/search/issues?q=is:pr+is:open+repo:%s&per_page=%d", repo, maxPRs), nil)
	issuesRequest, _ := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/search/issues?q=is:issue+is:open+repo:%s&per_page=%d", repo, maxIssues), nil)
	CommitsRequest, _ := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/commits?per_page=%d", repo, maxCommits), nil)

	if err != nil {
		return repository{
			Name: repo,
		}, fmt.Errorf("%w: could not create request with repository: %v", errNoContent, err)
	}

	if token != "" {

		token = fmt.Sprintf("Bearer %s", token)
		repositoryRequest.Header.Add("Authorization", token)
		PRsRequest.Header.Add("Authorization", token)
		issuesRequest.Header.Add("Authorization", token)
		CommitsRequest.Header.Add("Authorization", token)
	}

	var repositoryResponse githubRepositoryResponseJson
	var detailsErr error
	var PRsResponse githubTicketResponseJson
	var PRsErr error
	var issuesResponse githubTicketResponseJson
	var issuesErr error
	var commitsResponse []gitHubCommitResponseJson
	var CommitsErr error
	var wg sync.WaitGroup

	wg.Add(1)
	go (func() {
		defer wg.Done()
		repositoryResponse, detailsErr = decodeJsonFromRequest[githubRepositoryResponseJson](defaultHTTPClient, repositoryRequest)
	})()

	if maxPRs > 0 {
		wg.Add(1)
		go (func() {
			defer wg.Done()
			PRsResponse, PRsErr = decodeJsonFromRequest[githubTicketResponseJson](defaultHTTPClient, PRsRequest)
		})()
	}

	if maxIssues > 0 {
		wg.Add(1)
		go (func() {
			defer wg.Done()
			issuesResponse, issuesErr = decodeJsonFromRequest[githubTicketResponseJson](defaultHTTPClient, issuesRequest)
		})()
	}

	if maxCommits > 0 {
		wg.Add(1)
		go (func() {
			defer wg.Done()
			commitsResponse, CommitsErr = decodeJsonFromRequest[[]gitHubCommitResponseJson](defaultHTTPClient, CommitsRequest)
		})()
	}

	wg.Wait()

	if detailsErr != nil {
		return repository{}, fmt.Errorf("%w: could not get repository details: %s", errNoContent, detailsErr)
	}

	details := repository{
		Name:         repositoryResponse.Name,
		Stars:        repositoryResponse.Stars,
		Forks:        repositoryResponse.Forks,
		PullRequests: make([]githubTicket, 0, len(PRsResponse.Tickets)),
		Issues:       make([]githubTicket, 0, len(issuesResponse.Tickets)),
		Commits:      make([]githubCommitDetails, 0, len(commitsResponse)),
	}

	err = nil

	if maxPRs > 0 {
		if PRsErr != nil {
			err = fmt.Errorf("%w: could not get PRs: %s", errPartialContent, PRsErr)
		} else {
			details.OpenPullRequests = PRsResponse.Count

			for i := range PRsResponse.Tickets {
				details.PullRequests = append(details.PullRequests, githubTicket{
					Number:    PRsResponse.Tickets[i].Number,
					CreatedAt: parseRFC3339Time(PRsResponse.Tickets[i].CreatedAt),
					Title:     PRsResponse.Tickets[i].Title,
				})
			}
		}
	}

	if maxIssues > 0 {
		if issuesErr != nil {
			// TODO: fix, overwriting the previous error
			err = fmt.Errorf("%w: could not get issues: %s", errPartialContent, issuesErr)
		} else {
			details.OpenIssues = issuesResponse.Count

			for i := range issuesResponse.Tickets {
				details.Issues = append(details.Issues, githubTicket{
					Number:    issuesResponse.Tickets[i].Number,
					CreatedAt: parseRFC3339Time(issuesResponse.Tickets[i].CreatedAt),
					Title:     issuesResponse.Tickets[i].Title,
				})
			}
		}
	}

	if maxCommits > 0 {
		if CommitsErr != nil {
			err = fmt.Errorf("%w: could not get commits: %s", errPartialContent, CommitsErr)
		} else {
			for i := range commitsResponse {
				details.Commits = append(details.Commits, githubCommitDetails{
					Sha:       commitsResponse[i].Sha,
					Author:    commitsResponse[i].Commit.Author.Name,
					CreatedAt: parseRFC3339Time(commitsResponse[i].Commit.Author.Date),
					Message:   strings.SplitN(commitsResponse[i].Commit.Message, "\n\n", 2)[0],
				})
			}
		}
	}

	return details, err
}
