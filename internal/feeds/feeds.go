package feeds

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Beesy23/gator/internal/commands"
	"github.com/Beesy23/gator/internal/database"
	"github.com/google/uuid"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", "gator")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	feed := RSSFeed{}
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, err
	}

	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for i, item := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(item.Title)
		feed.Channel.Item[i].Description = html.UnescapeString(item.Description)
	}

	return &feed, nil
}

func HandlerAgg(s *commands.State, cmd commands.Command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <time_between_requests>", cmd.Name)
	}

	timeStr := cmd.Args[0]
	timeBetweenRequests, err := time.ParseDuration(timeStr)
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	fmt.Printf("Collecting feeds every %s\n", timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)
	defer ticker.Stop()

	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func scrapeFeeds(s *commands.State) error {

	ctx := context.Background()
	nextFeed, err := s.Db.GetNextFeedToFetch(ctx)
	if err != nil {
		return err
	}

	err = s.Db.MarkFeedFetched(ctx, nextFeed.ID)
	if err != nil {
		return err
	}

	feed, err := FetchFeed(ctx, nextFeed.Url)
	if err != nil {
		return err
	}

	for _, item := range feed.Channel.Item {
		published, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			return err
		}

		params := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: true},
			PublishedAt: sql.NullTime{Time: published, Valid: true},
			FeedID:      nextFeed.ID,
		}

		_, err = s.Db.CreatePost(ctx, params)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				continue
			}
			return err
		}
	}

	return nil
}

func HandlerAddFeed(s *commands.State, cmd commands.Command, user database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("usage: %s <name> <url>", cmd.Name)
	}

	userID := user.ID
	name := cmd.Args[0]
	feedURL := cmd.Args[1]

	params := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      name,
		Url:       feedURL,
		UserID:    userID,
	}
	feed, err := s.Db.CreateFeed(context.Background(), params)
	if err != nil {
		return fmt.Errorf("couldn't create feed: %w", err)
	}

	feedFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    userID,
		FeedID:    feed.ID,
	}

	_, err = s.Db.CreateFeedFollow(context.Background(), feedFollowParams)
	if err != nil {
		return err
	}

	fmt.Println("Feed created successfully:")
	printFeed(feed)

	return nil
}

func HandlerFeeds(s *commands.State, cmd commands.Command, user database.User) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: %s", cmd.Name)
	}

	feeds, err := s.Db.GetFeeds(context.Background())
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		fmt.Printf("%s: %s by %s\n", feed.Name, feed.Url, feed.UserName)
	}

	return nil
}

func HandlerFollow(s *commands.State, cmd commands.Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <url>", cmd.Name)
	}

	feed, err := s.Db.GetFeedFromURL(context.Background(), cmd.Args[0])
	if err != nil {
		return err
	}

	params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed.FeedID,
	}

	followed_feed, err := s.Db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return err
	}

	fmt.Printf("%s is now following feed: %s\n", followed_feed.FeedName, followed_feed.UserName)
	return nil
}

func HandlerUnfollow(s *commands.State, cmd commands.Command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <url>", cmd.Name)
	}

	feedURL := cmd.Args[0]
	feed, err := s.Db.GetFeedFromURL(context.Background(), feedURL)
	if err != nil {
		return err
	}
	params := database.DeleteFeedParams{
		FeedID: feed.FeedID,
		UserID: user.ID,
	}
	if err := s.Db.DeleteFeed(context.Background(), params); err != nil {
		return err
	}
	fmt.Printf("Successfully unfollowed feed: %s", feed.FeedName)
	return err
}

func HandlerFollowing(s *commands.State, cmd commands.Command, user database.User) error {
	if len(cmd.Args) != 0 {
		return fmt.Errorf("usage: %s", cmd.Name)
	}

	feedFollows, err := s.Db.GetFeedFollowsFromUser(context.Background(), user.ID)
	if err != nil {
		return err
	}

	for _, feed_name := range feedFollows {
		fmt.Println(feed_name)
	}
	return nil
}

func printFeed(feed database.Feed) {
	fmt.Printf("* ID:            %s\n", feed.ID)
	fmt.Printf("* Created:       %v\n", feed.CreatedAt)
	fmt.Printf("* Updated:       %v\n", feed.UpdatedAt)
	fmt.Printf("* Name:          %s\n", feed.Name)
	fmt.Printf("* URL:           %s\n", feed.Url)
	fmt.Printf("* UserID:        %s\n", feed.UserID)
	fmt.Println()
}
