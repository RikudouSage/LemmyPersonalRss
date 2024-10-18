package feed

import (
	"LemmyPersonalRss/cache"
	"LemmyPersonalRss/config"
	"LemmyPersonalRss/dto"
	"LemmyPersonalRss/helper"
	"fmt"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gorilla/feeds"
	"github.com/microcosm-cc/bluemonday"
	"io"
	"net/http"
	"strconv"
	"time"
)

func CreateFeedFromPosts(posts []*dto.LemmyPostView, appUser *dto.AppUser, page int, cachePool cache.ItemPool) *feeds.Feed {
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
	for _, post := range posts {
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
			if item.Enclosure == nil {
				// we know there's a link and for whatever reason it's not a valid image
				item.Content = fmt.Sprintf(
					"<a href=\"%s\" target=\"_blank\">The post contains a link, click here to visit it.</a><hr>%s",
					*post.Post.Url,
					item.Content,
				)
			}
		}

		feed.Items = append(feed.Items, item)
	}

	return feed
}
