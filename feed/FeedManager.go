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
	urlPkg "net/url"
	"strconv"
	"time"
)

func getHostFromUrl(url string) string {
	parsed, err := urlPkg.Parse(url)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return parsed.Host
}

func markdownToHtml(content string) string {
	markdownParser := parser.NewWithExtensions(parser.CommonExtensions)
	document := markdownParser.Parse([]byte(content))
	htmlRenderer := html.NewRenderer(html.RendererOptions{Flags: html.CommonFlags})
	return string(markdown.Render(document, htmlRenderer))
}

func htmlToPlain(content string) string {
	descriptionPolicy := bluemonday.StripTagsPolicy()
	return descriptionPolicy.Sanitize(content)
}

func CreateFeedForUser(appUser *dto.AppUser, page int, cachePool cache.ItemPool) *feeds.Feed {
	var instance string
	if appUser.Instance != nil {
		instance = *appUser.Instance
	} else {
		instance = config.GlobalConfiguration.Instance
	}
	now := time.Now()

	feed := &feeds.Feed{
		Title: fmt.Sprintf("Lemmy - @%s@%s saved list", appUser.Username, instance),
		Link: &feeds.Link{
			Href: fmt.Sprintf(
				"https://%s/u/%s?page=%d&sort=New&view=Saved",
				instance,
				appUser.Username,
				page,
			),
		},
		Description: fmt.Sprintf("Personal RSS feed created from saved posts by @%s@%s", appUser.Username, instance),
		Author: &feeds.Author{
			Name: fmt.Sprintf("@%s@%s", appUser.Username, instance),
		},
		Created: now,
		Image:   nil,
	}

	if appUser.ImageUrl != nil {
		feed.Image = &feeds.Image{
			Url:   *appUser.ImageUrl,
			Title: fmt.Sprintf("Lemmy - @%s@%s", appUser.Username, instance),
			Link:  fmt.Sprintf("https://%s/u/%s", instance, appUser.Username),
		}
	}

	return feed
}

func AddCommentsToFeed(feed *feeds.Feed, comments []*dto.LemmyCommentView, appUser *dto.AppUser) {
	var instance string
	if appUser.Instance != nil {
		instance = *appUser.Instance
	} else {
		instance = config.GlobalConfiguration.Instance
	}

	if instance == "" {
		return
	}

	for _, comment := range comments {
		item := &feeds.Item{
			Title: "Comment on post \"" + comment.Post.Name + "\"",
			Link: &feeds.Link{
				Href: fmt.Sprintf("https://%s/comment/%s", instance, comment.Comment.Id),
			},
			Author: &feeds.Author{
				Name: fmt.Sprintf("@%s@%s", comment.Creator.Name, getHostFromUrl(comment.Creator.ActorId)),
			},
			Description: htmlToPlain(markdownToHtml(comment.Comment.Content)),
			Created:     comment.Comment.Published.Time,
			Content:     markdownToHtml(comment.Comment.Content),
		}
		item.Id = item.Link.Href
		if len(item.Description) > 400 {
			item.Description = item.Description[:400] + "..."
		}

		feed.Items = append(feed.Items, item)
	}
}

func AddPostsToFeed(feed *feeds.Feed, posts []*dto.LemmyPostView, cachePool cache.ItemPool, appUser *dto.AppUser) {
	var instance string
	if appUser.Instance != nil {
		instance = *appUser.Instance
	} else {
		instance = config.GlobalConfiguration.Instance
	}

	for _, post := range posts {
		item := &feeds.Item{
			Title: post.Post.Name,
			Link: &feeds.Link{
				Href: fmt.Sprintf("https://%s/post/%d", instance, post.Post.Id),
			},
			Author: &feeds.Author{
				Name: fmt.Sprintf("@%s@%s", post.Creator.Name, getHostFromUrl(post.Creator.ActorId)),
			},
			Created: post.Post.Published.Time,
		}
		item.Id = item.Link.Href
		if post.Post.Updated != nil {
			item.Updated = post.Post.Updated.Time
		}
		if post.Post.Body != nil && *post.Post.Body != "" {
			rendered := markdownToHtml(*post.Post.Body)
			sanitizedDescription := htmlToPlain(rendered)

			item.Description = sanitizedDescription
			if len(item.Description) > 400 {
				item.Description = item.Description[:400] + "..."
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
}
