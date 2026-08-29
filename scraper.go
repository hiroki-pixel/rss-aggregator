package main

import (
	"context"
	"log"
	"rss-aggregator/internal/database"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

func scrapeFeed(wg *sync.WaitGroup, apiCfg *apiConfig, feed database.Feed) {
	defer wg.Done()

	log.Printf("Starting fetch: %s", feed.Name)

	rssFeed, err := urlToFeed(feed.Url)
	if err != nil {
		log.Printf("Couldn't get fetch feed %s: %v", feed.Name, err)
		return
	}

	for _, item := range rssFeed.Channel.Item {
		publishedAt, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			log.Printf("Couldn't parse published date:%v", err)
			continue
		}

		postParams := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			Description: item.Description,
			PublishedAt: publishedAt,
			FeedID:      feed.ID,
		}

		_, err = apiCfg.DB.CreatePost(
			context.Background(),
			postParams,
		)

		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				continue
			}

			log.Printf("Couldn't create post: %v", err)
			continue
		}
	}

	err = apiCfg.DB.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		log.Printf("Couldn't mark feed fetched:%v", err)
		return
	}
}

func startScraping(apiCfg *apiConfig, interval time.Duration, concurrency int32) {
	ticker := time.NewTicker(interval)

	defer ticker.Stop()

	for {
		feeds, err := apiCfg.DB.GetNextFeedsToFetch(context.Background(), concurrency)
		if err != nil {
			log.Printf("Couldn't get next feed to fetch:%v", err)
			<-ticker.C
			continue
		}

		var wg sync.WaitGroup

		for _, feed := range feeds {
			wg.Add(1)

			go scrapeFeed(&wg, apiCfg, feed)
		}

		wg.Wait()

		<-ticker.C
	}
}
