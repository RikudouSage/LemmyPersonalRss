package main

import (
	"LemmyPersonalRss/cache"
	"LemmyPersonalRss/config"
	. "LemmyPersonalRss/controller"
	"LemmyPersonalRss/database"
	"LemmyPersonalRss/database/migration"
	"LemmyPersonalRss/lemmy"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
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
	const feedUrl string = "https://{instance}/" + feedPath
	api := &lemmy.Api{
		Cache: cachePool,
	}

	http.HandleFunc("GET /rss/init", func(writer http.ResponseWriter, request *http.Request) {
		replacedUrl := strings.Replace(feedUrl, "{instance}", config.GlobalConfiguration.Instance, -1)
		HandleInit(writer, request, replacedUrl, db)
	})
	http.HandleFunc("GET /"+feedPath, func(writer http.ResponseWriter, request *http.Request) {
		HandleRssFeed(writer, request, feedPath, db, cachePool, api)
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
