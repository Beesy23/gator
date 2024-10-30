package posts

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Beesy23/gator/internal/commands"
	"github.com/Beesy23/gator/internal/database"
)

func HandlerBrowse(s *commands.State, cmd commands.Command, user database.User) error {
	limit := 2
	ctx := context.Background()
	if len(cmd.Args) == 1 {
		if specifiedLimit, err := strconv.Atoi(cmd.Args[0]); err == nil {
			limit = specifiedLimit
		} else {
			return fmt.Errorf("invalid limit: %w", err)
		}
	}

	params := database.GetPostsParams{
		UserID: user.ID,
		Limit:  int32(limit),
	}

	posts, err := s.Db.GetPosts(ctx, params)
	if err != nil {
		return fmt.Errorf("couldn't get posts for user: %w", err)
	}

	fmt.Printf("Found %d posts for user %s:\n", len(posts), user.Name)
	for _, post := range posts {
		fmt.Printf("%s from %s\n", post.PublishedAt.Format("Mon Jan 2"), post.FeedName)
		fmt.Printf("--- %s ---\n", post.Title)
		fmt.Printf("    %v\n", post.Description.String)
		fmt.Printf("Link: %s\n", post.Url)
		fmt.Println("=====================================")
	}

	return nil
}
