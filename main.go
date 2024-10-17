package main

import (
	"LemmyPersonalRss/cache"
	"LemmyPersonalRss/config"
	"LemmyPersonalRss/database"
	"LemmyPersonalRss/database/migration"
	"LemmyPersonalRss/helper"
	"LemmyPersonalRss/lemmy"
	"LemmyPersonalRss/user"
	"encoding/json"
	"fmt"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gorilla/feeds"
	"github.com/microcosm-cc/bluemonday"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	gracefulShutdown := make(chan os.Signal, 1)
	signal.Notify(gracefulShutdown, syscall.SIGINT, syscall.SIGTERM)

	var err error
	var db database.Database
	var cachePool cache.ItemPool = &cache.InMemoryCacheItemPool{}

	if config.GlobalConfiguration.DatabasePath == nil {
		db = &database.InMemoryDatabase{}
	} else {
		db, err = database.NewSqliteDatabase(*config.GlobalConfiguration.DatabasePath, migration.GetManager())
		if err != nil {
			panic(err)
		}
	}

	const feedPath string = "rss/{hash}"
	feedUrl := "https://" + config.GlobalConfiguration.Instance + "/" + feedPath
	api := &lemmy.Api{
		Cache: cachePool,
	}

	http.HandleFunc("GET /rss/init", func(writer http.ResponseWriter, request *http.Request) {
		if config.GlobalConfiguration.Logging {
			fmt.Println("GET /rss/init called")
			defer func() {
				fmt.Println("GET /rss/init finished")
			}()
		}

		writer.Header().Set("Content-Type", "application/json")

		currentUser := user.GetCurrentFromHttpContext(request, db)
		if currentUser == nil {
			currentUser = user.CreateFromHttpContext(request, db)
		}

		if currentUser == nil {
			response := map[string]string{
				"error": "Failed to get current user. Are you logged in?",
			}
			raw, err := json.Marshal(response)
			if err != nil {
				fmt.Println(err)
			}

			writer.WriteHeader(http.StatusUnauthorized)
			_, err = writer.Write(raw)
			if err != nil {
				fmt.Println(err)
			}

			if config.GlobalConfiguration.Logging {
				fmt.Println("User is not logged in")
			}
			return
		}

		url := strings.Replace(feedUrl, "{hash}", currentUser.Hash, -1)
		response := map[string]string{
			"message": "Success! You can find your feed at " + url,
			"url":     url,
		}

		raw, err := json.Marshal(response)
		if err != nil {
			fmt.Println(err)
		}
		_, err = writer.Write(raw)
		if err != nil {
			fmt.Println(err)
		}
	})
	http.HandleFunc("GET /"+feedPath, func(writer http.ResponseWriter, request *http.Request) {
		urlHash := request.PathValue("hash")
		if config.GlobalConfiguration.Logging {
			fmt.Println(strings.Replace("GET /"+feedPath+" called", "{hash}", urlHash, -1))
			defer func() {
				fmt.Println(strings.Replace("GET /"+feedPath+" finished", "{hash}", urlHash, -1))
			}()
		}

		writer.Header().Set("Content-Type", "application/json")

		appUser := db.FindByHash(urlHash)

		if appUser == nil {
			response := map[string]string{
				"error": "The RSS feed could not be found.",
			}
			raw, err := json.Marshal(response)
			if err != nil {
				fmt.Println(err)
			}

			writer.WriteHeader(http.StatusNotFound)
			_, err = writer.Write(raw)
			if err != nil {
				fmt.Println(err)
			}

			if config.GlobalConfiguration.Logging {
				fmt.Println("RSS feed not found")
			}
			return
		}
		go func() {
			cacheItem := cachePool.Get(fmt.Sprintf("%s.%d", "user_refresh_blocker", appUser.Id))
			if cacheItem.Hit() {
				return
			}

			err := user.UpdateUserData(appUser, db)
			if err != nil {
				fmt.Println(err)
			}
			cacheItem.Set(true)
			cacheItem.SetExpiresAfter(&config.GlobalConfiguration.CacheDuration)
			err = cachePool.Store(cacheItem)
			if err != nil {
				fmt.Println(err)
			}
		}()

		page := (func() int {
			raw := request.URL.Query().Get("page")
			if raw == "" {
				return 1
			}

			parsed, err := strconv.Atoi(raw)
			if err != nil {
				fmt.Println(err)
				return 1
			}

			return parsed
		})()
		perPage := (func() int {
			raw := request.URL.Query().Get("limit")
			if raw == "" {
				return 20
			}

			parsed, err := strconv.Atoi(raw)
			if err != nil {
				fmt.Println(err)
				return 20
			}

			return parsed
		})()

		saved := api.GetSavedPosts(appUser, page, perPage)

		now := time.Now()
		feed := &feeds.Feed{
			Title: fmt.Sprintf("Lemmy - @%s@%s saved list", appUser.Username, config.GlobalConfiguration.Instance),
			Link: &feeds.Link{
				Href: fmt.Sprintf(
					"https://%s/u/%s?page=%d&sort=New&view=Saved",
					config.GlobalConfiguration.Instance,
					appUser.Username,
					page,
				),
			},
			Description: fmt.Sprintf("Personal RSS feed created from saved posts by @%s@%s", appUser.Username, config.GlobalConfiguration.Instance),
			Author: &feeds.Author{
				Name: fmt.Sprintf("@%s@%s", appUser.Username, config.GlobalConfiguration.Instance),
			},
			Created: now,
			Image:   nil,
		}
		if appUser.ImageUrl != nil {
			feed.Image = &feeds.Image{
				Url:   *appUser.ImageUrl,
				Title: fmt.Sprintf("Lemmy - @%s@%s", appUser.Username, config.GlobalConfiguration.Instance),
				Link:  fmt.Sprintf("https://%s/u/%s", config.GlobalConfiguration.Instance, appUser.Username),
			}
		}
		for _, post := range saved {
			item := &feeds.Item{
				Title: post.Post.Name,
				Link: &feeds.Link{
					Href: fmt.Sprintf("https://%s/post/%d", config.GlobalConfiguration.Instance, post.Post.Id),
				},
				Author: &feeds.Author{
					Name: fmt.Sprintf("@%s@%s", post.Creator.Name, config.GlobalConfiguration.Instance),
				},
				Created: post.Post.Published.Time,
			}
			item.Id = item.Link.Href
			if post.Post.Updated != nil {
				item.Updated = post.Post.Updated.Time
			}
			if post.Post.Body != nil && *post.Post.Body != "" {
				markdownParser := parser.NewWithExtensions(parser.CommonExtensions)
				document := markdownParser.Parse([]byte(*post.Post.Body))
				htmlRenderer := html.NewRenderer(html.RendererOptions{Flags: html.CommonFlags})
				rendered := string(markdown.Render(document, htmlRenderer))
				descriptionPolicy := bluemonday.StripTagsPolicy()
				sanitizedDescription := descriptionPolicy.Sanitize(rendered)

				item.Description = sanitizedDescription
				if len(item.Description) > 400 {
					item.Description = sanitizedDescription[:400] + "..."
				}
				item.Content = rendered
			}

			imgExtensionMimeTypeMap := map[string]string{
				".jpg":  "image/jpeg",
				".jpeg": "image/jpeg",
				".png":  "image/png",
				".gif":  "image/gif",
				".webp": "image/webp",
				".svg":  "image/svg+xml",
			}
			if post.Post.Url != nil && *post.Post.Url != "" {
				item.Enclosure = (func() *feeds.Enclosure {
					if !helper.EndsWithAny(*post.Post.Url, helper.Keys(imgExtensionMimeTypeMap)) {
						return nil
					}

					contentType := imgExtensionMimeTypeMap["."+helper.ExtractExtension(*post.Post.Url)]
					cacheItem := cachePool.Get(fmt.Sprintf("%s.%s", "image_size", *post.Post.Url))
					if cacheItem.Hit() {
						return &feeds.Enclosure{
							Url:    *post.Post.Url,
							Length: cacheItem.Get().(string),
							Type:   contentType,
						}
					}

					httpClient := &http.Client{}
					response, err := httpClient.Get(*post.Post.Url)
					if err != nil {
						fmt.Println(err)
						return nil
					}
					defer response.Body.Close()

					var length int64
					if response.ContentLength > -1 {
						length = response.ContentLength
					} else {
						body, err := io.ReadAll(response.Body)
						if err != nil {
							fmt.Println(err)
							return nil
						}
						length = int64(len(body))
					}

					lengthString := strconv.FormatInt(length, 10)
					cacheItem.Set(lengthString)
					err = cachePool.Store(cacheItem)
					if err != nil {
						fmt.Println(err)
					}

					return &feeds.Enclosure{
						Url:    *post.Post.Url,
						Length: lengthString,
						Type:   contentType,
					}
				})()

			}

			feed.Items = append(feed.Items, item)
		}
		rss, err := feed.ToRss()
		if err != nil {
			fmt.Println(err)

			response := map[string]string{
				"error": "The RSS feed could not be generated, please try again later.",
			}
			raw, err := json.Marshal(response)
			if err != nil {
				fmt.Println(err)
			}

			writer.WriteHeader(http.StatusInternalServerError)
			_, err = writer.Write(raw)
			if err != nil {
				fmt.Println(err)
			}

			return
		}

		writer.Header().Set("Content-Type", "application/rss+xml")
		_, err = writer.Write([]byte(rss))
		if err != nil {
			fmt.Println(err)
			return
		}
	})

	go func() {
		fmt.Println("Starting server at port", config.GlobalConfiguration.Port)
		err := http.ListenAndServe(fmt.Sprintf(":%d", config.GlobalConfiguration.Port), nil)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	<-gracefulShutdown
	fmt.Println("Shutting down server...")
}
