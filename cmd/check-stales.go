package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/KhashayarKhm/anti-stale/github"
	"github.com/KhashayarKhm/anti-stale/utils"
	"github.com/spf13/cobra"
)

type CheckStaleEntitiesOpts struct {
	Reply      bool
	Intractive bool
	Msg        string
	Label      string
}

func init() {
	checkCmd := &cobra.Command{
		Use:   "check",
		Short: "find stale issue/pr(s)",
		Run: func(c *cobra.Command, _ []string) {
			reply, _ := c.Flags().GetBool("reply")
			intr, _ := c.Flags().GetBool("interactive")
			msg, _ := c.Flags().GetString("msg")
			label, _ := c.Flags().GetString("label")

			opts := &CheckStaleEntitiesOpts{
				Reply:      reply,
				Intractive: intr,
				Msg:        msg,
				Label:      label,
			}

			checkStaleEntities(globalConfig, opts)
		},
	}

	checkCmd.Flags().Bool("reply", false, "Reply the issue/pr(s) that is stale")
	checkCmd.Flags().BoolP("interactive", "i", false, "Make decision to reply or not for every issue/pr(s) that is stale")
	checkCmd.Flags().String("msg", "not stale", "The message to reply")
	checkCmd.Flags().StringP("label", "l", "Stale", "Stale label name")

	rootCmd.AddCommand(checkCmd)
}

func checkStaleEntities(config *utils.Config, opts *CheckStaleEntitiesOpts) {
	ghc, err := github.Init("https://api.github.com/graphql", config.Token, config.UserAgent)
	if err != nil {
		logger.Fatal(err)
	}

	res, err := ghc.ListEntitiesByRepo(context.Background(), logger, &config.Owners)
	if errors.Is(err, utils.ErrGraphQLResponseErrors) {
		for _, e := range res.Errors {
			logger.Error(fmt.Sprintf("%v: %v", utils.ErrGraphQLResponseErrors, e))
		}
	} else if err != nil {
		logger.Fatal(err)
	}

	staledEntities := make([]github.GhEntity, 0, 10)
	for _, entities := range res.Data {
		for _, entVal := range entities {
			staled := false
			logger.Debug(fmt.Sprintf("start checking %s (Closed: %t)", entVal.URL, entVal.Closed))
			if !entVal.Closed {
				logger.Debug(fmt.Sprintf("%s labels: %v", entVal.URL, entVal.Labels.Nodes))
				for _, l := range entVal.Labels.Nodes {
					if l.Name == opts.Label {
						staled = true
						break
					}
				}

				if staled {
					logger.Info(fmt.Sprintf("%s is stale", entVal.URL))
					if opts.Reply {
						if opts.Intractive {
							var answer string
							fmt.Printf("Do I reply to this issue/pr? (%s) [Y/n]: ", entVal.URL)
							fmt.Scanln(&answer)
							if answer == "n" || answer == "no" || answer == "No" {
								continue
							}
						}
						staledEntities = append(staledEntities, entVal)
					}
				} else {
					logger.Info(fmt.Sprintf("%s is not stale", entVal.URL))
				}
			} else {
				logger.Info(fmt.Sprintf("%s is not open, skipping", entVal.URL))
			}
		}
	}

	if opts.Reply && len(staledEntities) != 0 {
		res, err := ghc.LeaveCommentOnEntities(context.Background(), logger, opts.Msg, &staledEntities)
		if errors.Is(err, utils.ErrGraphQLResponseErrors) {
			for _, e := range res.Errors {
				logger.Error(fmt.Sprintf("%v: %v", utils.ErrGraphQLResponseErrors, e))
			}
		} else if err != nil {
			logger.Fatal(err)
		}

		for _, r := range res.Data {
			logger.Info(fmt.Sprintf("leave comment successfuly: %s", r.CommentEdge.URL))
		}
	}
}
