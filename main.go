package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/deoreal/gator/internal/config"
	"github.com/deoreal/gator/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

const configFileName = ".gatorconfig.json"

var time_between_reqs = "10s"

type state struct {
	conf *config.Config
	db   *database.Queries
}

type command struct {
	name string
	args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

type Feed struct {
	ID        int       `json:"title"`
	CreatedAt time.Time `json:"create_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	UserID    uuid.UUID `json:"user_id"`
}

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

// fetchFeed reads a RSSfeed from a given url
func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	client := &http.Client{
		CheckRedirect: http.DefaultClient.CheckRedirect,
	}

	req, err := http.NewRequest("GET", feedURL, nil)
	if err != nil {
		return &RSSFeed{}, err
	}

	req.Header.Set("User-Agent", "gator")
	resp, err := client.Do(req)
	if err != nil {
		return &RSSFeed{}, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &RSSFeed{}, err
	}

	xmldata := RSSFeed{}

	err = xml.Unmarshal(body, &xmldata)
	if err != nil {
		return &RSSFeed{}, err
	}

	xmldata.Channel.Title = html.UnescapeString(xmldata.Channel.Title)
	xmldata.Channel.Description = html.UnescapeString((xmldata.Channel.Description))

	return &xmldata, nil
}

// middlewareLoggedIn used to enrich a handler call with needed information
func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		// Check if a user is currently logged in
		if s.conf.CurrentUserName == "" {
			return fmt.Errorf("no user is currently logged in")
		}

		// Get the user ID from the database
		userID, err := s.db.GetUser(context.Background(), s.conf.CurrentUserName)
		if err != nil {
			return fmt.Errorf("couldn't get current user: %w", err)
		}

		// Get the full user object
		user, err := s.db.GetUserByID(context.Background(), userID)
		if err != nil {
			return fmt.Errorf("couldn't get user details: %w", err)
		}

		// Call the wrapped handler with the user
		return handler(s, cmd, user)
	}
}

// getConfigFilePath returns the path of the config fiel namt
func getConfigFilePath() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to read homedir: %s", err)
	}
	return homedir + "/" + configFileName, nil
}

// handlerLogin set a user as current user
func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		fmt.Println("username is required")
		os.Exit(1)
	}

	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	if slices.Contains(users, cmd.args[0]) {

		s.conf.CurrentUserName = cmd.args[0]
		config.WriteConfig(*s.conf)
		fmt.Println("user has been set to: ", cmd.args[0])
	} else {
		return fmt.Errorf("unknown user")
	}
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}
	fmt.Println(feeds)
	return nil
}

// handlerUsers lists the list of registeres users and indicates which is set as the current user
func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		if user == s.conf.CurrentUserName {
			fmt.Println(user, "(current)")
		} else {
			fmt.Println(user)
		}
	}

	return nil
}

// handlerRegister registers a new user by adding him to the database and setting it as the current user
func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		fmt.Println("username is required")
		os.Exit(1)

	}

	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	if slices.Contains(users, cmd.args[0]) {
		return fmt.Errorf("user %s already in database", cmd.args[0])
	}
	//	fmt.Println(cmd.args)
	err = s.db.CreateUser(context.Background(), database.CreateUserParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: cmd.args[0]})
	if err != nil {
		return err
	}
	fmt.Printf("user %s has been added to the database\n", cmd.args[0])

	s.conf.CurrentUserName = cmd.args[0]

	s.conf.SetUser(cmd.args[0])
	config.WriteConfig(*s.conf)
	return nil
}

// handlerAgg lists an RSSfeed target
func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) == 1 {
		time_between_reqs = cmd.args[0]
	}

	t, err := time.ParseDuration(time_between_reqs)
	if err != nil {
		return err
	}
	fmt.Println("Collecting feeds every:", t)
	feedURL := "https://www.wagslane.dev/index.xml"

	rssFeed, err := fetchFeed(context.Background(), feedURL)
	if err != nil {
		return err
	}

	fmt.Println(rssFeed)
	timeBetweenRequests, _ := time.ParseDuration(time_between_reqs)
	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

// handlerFollowing lists the feeds a user is assigned to
func handlerFollowing(s *state, cmd command, user database.User) error {
	feedNames, err := s.db.GetFeedFollowsForUser(context.Background(), user.Name)
	if err != nil {
		return fmt.Errorf("couldn't get feeds for user: %w", err)
	}

	if len(feedNames) == 0 {
		fmt.Println("No feeds found for the current user")
		return nil
	}

	fmt.Printf("Feeds followed by %s:\n", user.Name)
	for _, feedName := range feedNames {
		if feedName.Valid {
			fmt.Printf("- %s\n", feedName.String)
		}
	}

	return nil
}

// handlerFollow add a user to the list of followers
func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		fmt.Println("feed url is required")
		os.Exit(1)
	}

	feedURL := sql.NullString{String: cmd.args[0], Valid: true}

	feed, err := s.db.GetFeed(context.Background(), feedURL)
	if err != nil {
		return err
	}

	cff, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{UserID: user.ID, FeedID: feed.ID})
	if err != nil {
		return err
	}
	fmt.Println(cff)
	return nil
}

// handlerAddFeed adds a feed to the feeds table
func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		fmt.Println("name and url are required")
		os.Exit(1)
	}

	n := sql.NullString{String: cmd.args[0], Valid: true}
	u := sql.NullString{String: cmd.args[1], Valid: true}

	// Create the feed and get the created feed back
	createdFeed, err := s.db.CreateFeed(context.Background(),
		database.CreateFeedParams{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      n,
			Url:       u,
			UserID:    user.ID,
		})
	if err != nil {
		return err
	}

	// Automatically create a feed follow for the current user
	feedFollow, err := s.db.CreateFeedFollow(context.Background(),
		database.CreateFeedFollowParams{
			UserID: user.ID,
			FeedID: createdFeed.ID,
		})
	if err != nil {
		return err
	}

	fmt.Printf("Feed created: %s\n", createdFeed.Name.String)
	fmt.Printf("User %s is now following %s\n", feedFollow.UserName, feedFollow.FeedName.String)
	return nil
}

// handlerUnfollow removes a user from following a feed by its URL
func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		fmt.Println("feed url is required")
		os.Exit(1)
	}

	feedURL := sql.NullString{String: cmd.args[0], Valid: true}

	// Get the feed by URL
	feed, err := s.db.GetFeed(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("couldn't find feed with URL %s: %w", cmd.args[0], err)
	}

	// Delete the feed follow relationship
	err = s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("couldn't unfollow feed: %w", err)
	}

	fmt.Printf("User %s has unfollowed %s\n", user.Name, feed.Name.String)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.Reset(context.Background())
	if err != nil {
		return fmt.Errorf("failed to truncate users table: %s", err)
	}
	fmt.Printf("truncated the users table: %v\n", cmd)
	return nil
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.handlers[cmd.name]
	if !ok {
		return fmt.Errorf("command not registered")
	}

	return handler(s, cmd)
}

func (c *commands) register(name string, f func(s *state, cmd command) error) {
	c.handlers[name] = f
}

func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get next feed: %w", err)
	}

	if err := s.db.MarkFeedFetched(context.Background(), feed.ID); err != nil {
		return fmt.Errorf("failed to mark feed %d as fetched: %w", feed.ID, err)
	}

	// Fetch the feed content
	feedContent, err := fetchFeed(context.Background(), feed.Url.String)
	if err != nil {
		return fmt.Errorf("failed to fetch feed %s: %w", feed.Url.String, err)
	}

	// Iterate over items and save them to database
	for _, item := range feedContent.Channel.Item {
		// Parse the published date
		var publishedAt sql.NullTime
		if item.PubDate != "" {
			// Try multiple date formats commonly used in RSS feeds
			formats := []string{
				time.RFC1123Z,
				time.RFC1123,
				"Mon, 2 Jan 2006 15:04:05 -0700",
				"Mon, 2 Jan 2006 15:04:05 MST",
				"2006-01-02T15:04:05Z07:00",
				"2006-01-02T15:04:05Z",
				"2006-01-02 15:04:05",
			}

			for _, format := range formats {
				if parsedTime, err := time.Parse(format, item.PubDate); err == nil {
					publishedAt = sql.NullTime{Time: parsedTime, Valid: true}
					break
				}
			}
		}

		// Create post params
		postParams := database.CreatePostParams{
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       sql.NullString{String: html.UnescapeString(item.Title), Valid: item.Title != ""},
			Url:         item.Link,
			Description: sql.NullString{String: html.UnescapeString(item.Description), Valid: item.Description != ""},
			PublishedAt: publishedAt,
			FeedID:      feed.ID,
		}

		// Try to create the post
		_, err := s.db.CreatePost(context.Background(), postParams)
		if err != nil {
			// If it's a unique constraint violation (URL already exists), ignore it
			if strings.Contains(err.Error(), "unique constraint") || strings.Contains(err.Error(), "duplicate key") {
				continue
			}
			// For other errors, log them but don't stop the process
			log.Printf("Failed to create post %s: %v", item.Link, err)
		}
	}

	return nil
}

// handlerBrowse shows posts for the current user
func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := int32(2) // default limit

	if len(cmd.args) > 0 {
		parsedLimit, err := strconv.Atoi(cmd.args[0])
		if err != nil {
			return fmt.Errorf("invalid limit: %w", err)
		}
		limit = int32(parsedLimit)
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  limit,
	})
	if err != nil {
		return fmt.Errorf("couldn't get posts for user: %w", err)
	}

	if len(posts) == 0 {
		fmt.Println("No posts found for the current user")
		return nil
	}

	fmt.Printf("Posts for %s:\n", user.Name)
	for _, post := range posts {
		fmt.Printf("Title: %s\n", post.Title.String)
		fmt.Printf("URL: %s\n", post.Url)
		if post.Description.Valid {
			// Limit description to first 100 characters for readability
			desc := post.Description.String
			if len(desc) > 100 {
				desc = desc[:100] + "..."
			}
			fmt.Printf("Description: %s\n", desc)
		}
		if post.PublishedAt.Valid {
			fmt.Printf("Published: %s\n", post.PublishedAt.Time.Format("2006-01-02 15:04:05"))
		}
		fmt.Printf("Feed: %s\n", post.FeedName.String)
		fmt.Println("---")
	}

	return nil
}

// handlerScrape runs a one-time scrape of all feeds
func handlerScrape(s *state, cmd command) error {
	fmt.Println("Starting one-time scrape of all feeds...")
	err := scrapeFeeds(s)
	if err != nil {
		return fmt.Errorf("failed to scrape feeds: %w", err)
	}
	fmt.Println("Scrape completed successfully!")
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("not enough arguments")
		os.Exit(1)
	}

	c := &commands{
		handlers: make(map[string]func(*state, command) error),
	}
	s := &state{}
	var err error

	s.conf, err = config.ReadConfig()
	if err != nil {
		fmt.Printf("failed to read config: %s\n", err)
	}
	c.register("login", handlerLogin)
	c.register("register", handlerRegister)
	c.register("reset", handlerReset)
	c.register("users", handlerUsers)
	c.register("agg", handlerAgg)
	c.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	c.register("feeds", handlerFeeds)
	c.register("follow", middlewareLoggedIn(handlerFollow))
	c.register("following", middlewareLoggedIn(handlerFollowing))
	c.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	c.register("browse", middlewareLoggedIn(handlerBrowse))
	c.register("scrape", handlerScrape)

	cmd := command{
		name: os.Args[1],
		args: os.Args[2:],
	}
	db, err := sql.Open("postgres", s.conf.DBURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	dbQueries := database.New(db)

	s.db = dbQueries

	if err = c.run(s, cmd); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
