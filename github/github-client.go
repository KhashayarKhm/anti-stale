package github

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/KhashayarKhm/anti-stale/utils"
)

var (
	ErrGhURLIsInvalid     = errors.New("URL is invalid")
	ErrGhUserAgentIsEmpty = errors.New("User agent is invalid")
)

type GhEntityType int

const (
	GhIssueType GhEntityType = iota
	GhPrType
)

type GhClient struct {
	Token     string
	UserAgent string
	gqlClient *utils.GQLC
}

type GhLabel struct {
	Nodes []GhLabelNode `json:"nodes"`
}

type GhLabelNode struct {
	Name string `json:"name"`
}

type GhAuthor struct {
	Login string `json:"login"`
}

// Github issue/pull request
type GhEntity struct {
	Closed bool    `json:"closed"`
	ID     string  `json:"id"`
	URL    string  `json:"url"`
	Labels GhLabel `json:"labels"`
}

type GhAddCommentInput struct {
	SubjectID string `json:"subjectId"`
	Body      string `json:"body"`
}

// Example:
//
//	{
//	  "<repo-id>": {
//	    "<entity-id>": {
//	      ...
//	    }
//	  }
//	}
type ListIssuesByRepoData map[string]map[string]GhEntity

type LeaveCommentOnEntitiesDataCommentEdgeNode struct {
	URL string `json:"url"`
}

type LeaveCommentOnEntitiesDataCommentEdge struct {
	CommentEdge LeaveCommentOnEntitiesDataCommentEdgeNode `json:"commentEdge"`
}

type LeaveCommentOnEntitiesData map[string]LeaveCommentOnEntitiesDataCommentEdge

func Init(url, token, userAgent string) (GhClient, error) {
	gh := GhClient{}
	if url == "" {
		return gh, ErrGhURLIsInvalid
	} else if userAgent == "" {
		return gh, ErrGhUserAgentIsEmpty
	}

	gh.Token = token
	gh.UserAgent = userAgent
	gh.gqlClient = &utils.GQLC{URL: url}

	return gh, nil
}

func (gh *GhClient) ListEntitiesByRepo(ctx context.Context, logger *utils.Log, owners *utils.Owners) (utils.GraphQLResponse[ListIssuesByRepoData], error) {
	q, vars := gh.buildGQLIssueQuery(owners)
	logger.Debug(fmt.Sprintf("graphql query: %s", q))
	logger.Debug(fmt.Sprintf("graphql variables: %v", vars))
	headers := gh.getHeaders()
	res, err := utils.SendGQLRequest[ListIssuesByRepoData](gh.gqlClient, ctx, q, vars, headers)
	if err != nil {
		return res, err
	}

	return res, nil
}

func (gh *GhClient) LeaveCommentOnEntities(ctx context.Context, logger *utils.Log, msg string, staledEntites *[]GhEntity) (utils.GraphQLResponse[LeaveCommentOnEntitiesData], error) {
	q, vars := gh.buildGQLCommentQuery(msg, staledEntites)
	logger.Debug(fmt.Sprintf("graphql query: %s", q))
	logger.Debug(fmt.Sprintf("graphql variables: %v", vars))
	headers := gh.getHeaders()
	res, err := utils.SendGQLRequest[LeaveCommentOnEntitiesData](gh.gqlClient, ctx, q, vars, headers)
	if err != nil {
		return res, err
	}
	return res, nil
}

func (gh *GhClient) getHeaders() map[string]string {
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", gh.Token),
		"User-Agent":    gh.UserAgent,
		"Content-Type":  "application/json",
		"Accept":        "application/vnd.github+json",
	}
}

func (gh *GhClient) buildGQLCommentQuery(msg string, staledEntites *[]GhEntity) (string, map[string]GhAddCommentInput) {
	variables := make(map[string]GhAddCommentInput)
	n := 0
	qBody := make([]string, 0, len(*staledEntites))

	for _, ent := range *staledEntites {
		inpID := fmt.Sprintf("n%d", n)
		n++
		variables[inpID] = GhAddCommentInput{SubjectID: ent.ID, Body: msg}
		acID := fmt.Sprintf("n%d", n)
		n++
		mutQuery := fmt.Sprintf("%s: AddComment(input: %s) { url }", acID, inpID)
		qBody = append(qBody, mutQuery)
	}

	vars := make([]string, len(variables))
	vsi := 0
	for k := range variables {
		vars[vsi] = fmt.Sprintf("$%s: AddCommentInput!", k)
		vsi++
	}

	base := `mutation(%s) {
		%s
	}`
	q := fmt.Sprintf(base, strings.Join(vars, ","), strings.Join(qBody, "\n"))

	return q, variables
}

func (gh *GhClient) buildGQLIssueQuery(owners *utils.Owners) (string, map[string]any) {
	variables := make(map[string]any)
	n := 0
	qBody := make([]string, 0, len(*owners))
	for owner, repos := range *owners {
		ownerID := fmt.Sprintf("n%d", n)
		n++
		variables[ownerID] = owner
		for repo, entities := range repos {
			entitiesQuery := make([]string, len(entities.Issues)+len(entities.PRs))
			eqi := 0
			repoID := fmt.Sprintf("n%d", n)
			n++
			variables[repoID] = repo

			for _, issue := range entities.Issues {
				issueID := fmt.Sprintf("n%d", n)
				n++
				variables[issueID] = issue
				entitiesQuery[eqi] = fmt.Sprintf(`
				%s: issue(number: $%s) {
					id
					closed
					url
					labels (first: $first) {
						nodes {
							name
						}
					}
				}`, issueID, issueID)
				eqi++
			}

			for _, pr := range entities.PRs {
				prID := fmt.Sprintf("n%d", pr)
				n++
				variables[prID] = pr
				entitiesQuery[eqi] = fmt.Sprintf(`
				%s: pullRequest(number: $%s) {
					id
					closed
					url
					labels (first: $first) {
						nodes {
							name
						}
					}
				}`, prID, prID)
				eqi++
			}

			entitiesQueryStr := strings.Join(entitiesQuery, "\n")
			repoQuery := fmt.Sprintf(`%s: repository(owner: $%s, name: $%s) { %s }`, repoID, ownerID, repoID, entitiesQueryStr)
			qBody = append(qBody, repoQuery)
		}
	}

	variables["first"] = 5

	vars := make([]string, len(variables))
	vsi := 0
	for k, v := range variables {
		switch v.(type) {
		case string:
			vars[vsi] = fmt.Sprintf("$%s: String!", k)
		case int:
			vars[vsi] = fmt.Sprintf("$%s: Int!", k)
		default:
		}
		vsi++
	}

	base := `query (%s) {
		%s
	}`

	q := fmt.Sprintf(base, strings.Join(vars, ","), strings.Join(qBody, "\n"))

	return q, variables
}
