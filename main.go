package main

import (
	"LemmyPersonalRss/cache"
	"LemmyPersonalRss/config"
	. "LemmyPersonalRss/controller"
	"LemmyPersonalRss/database"
	"LemmyPersonalRss/database/migration"
	"LemmyPersonalRss/lemmy"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
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
	cleaner := cachePool.GetCleaner()
	if cleaner != nil {
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()
			for {
				fmt.Println("Cache cleaner is running")
				cleaner.Clean()
				fmt.Println("Cache cleaner finished")
				<-ticker.C
			}
		}()
	}

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
		HandleInitEndpoint(writer, request, replacedUrl, db)
	})
	http.HandleFunc("POST /rss/register", func(writer http.ResponseWriter, request *http.Request) {
		HandleRegisterEndpoint(writer, request, api, db, feedUrl)
	})
	http.HandleFunc("GET /"+feedPath, func(writer http.ResponseWriter, request *http.Request) {
		HandleRssFeedEndpoint(writer, request, feedPath, db, cachePool, api)
	})
	http.HandleFunc("GET /rss/config", func(writer http.ResponseWriter, request *http.Request) {
		HandleConfigEndpoint(writer)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.GlobalConfiguration.Port),
		Handler: nil,
	}

	go func() {
		fmt.Println("Starting server at port", config.GlobalConfiguration.Port)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	<-gracefulShutdown
	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Println("Server forced to shutdown:", err)
	} else {
		fmt.Println("Server exited properly")
	}
}
