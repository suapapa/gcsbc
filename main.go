package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gokyle/filecache"
)

var (
	loc       = time.FixedZone("KST", +9*60*60)
	urlPrefix string
)

func main() {
	fExpire := flag.Int("e",
		filecache.DefaultExpireItem,
		"maximum number of seconds between accesses a file can stay in the cache")
	fGarbage := flag.Int("g", filecache.DefaultEvery,
		"scan the cache for expired items every <n> seconds")
	fItems := flag.Int("n",
		100, //filecache.DefaultMaxItems,
		"max number of files to store in the cache")
	fPort := flag.Int("p", 8080, "port to listen on")
	fSize := flag.Int64("s", filecache.DefaultMaxSize, "max file size to cache")
	fDumpCache := flag.String("d", "",
		"dump cache stats duration; by default, this is turned off. Must be parsable with time.ParseDuration.")
	flag.StringVar(&urlPrefix, "f", "", "remove prefix subpath in url.PATH")
	flag.Parse()

	if urlPrefix != "" && !strings.HasPrefix(urlPrefix, "/") {
		urlPrefix = "/" + urlPrefix
	}

	srvWD := "."
	if flag.NArg() > 0 {
		srvWD = flag.Arg(0)
	}

	srvWD, err := filepath.Abs(srvWD)
	if err != nil {
		log.Fatal(err)
	}

	cache := &filecache.FileCache{
		Root:       srvWD,
		MaxSize:    *fSize,
		MaxItems:   *fItems,
		ExpireItem: *fExpire,
		Every:      *fGarbage,
	}
	cache.Start()

	if *fDumpCache != "" {
		go func() {
			dur, err := time.ParseDuration(*fDumpCache)
			if err != nil {
				fmt.Printf("[-] couldn't parse %s: %v\n",
					*fDumpCache, err)
				return
			}
			tk := time.NewTicker(dur)
			defer tk.Stop()
			for range tk.C {
				displayCacheStats(cache)
			}
		}()
	}

	http.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			urlPath := r.URL.Path
			if !strings.HasPrefix(urlPath, "/") {
				urlPath = "/" + urlPath
			}

			firstURLPrefixEndIdx := strings.Index(urlPath[1:], "/") + 1
			var urlPathPre, urlPathSur string
			switch firstURLPrefixEndIdx {
			case 0, len(urlPath): // wholeURL is prefix
				urlPathPre = urlPath[:firstURLPrefixEndIdx]
				urlPathSur = ""
			default:
				urlPathPre = urlPath[:firstURLPrefixEndIdx]
				urlPathSur = urlPath[firstURLPrefixEndIdx:]
			}

			if urlPathSur == "" || strings.Compare(urlPrefix, urlPathPre) != 0 {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, "404")
				return
			}

			// log.Printf(
			// 	"urlPrefix=%s, urlPath=%s, urlPathPre=%s, urlPathSur=%s",
			// 	urlPrefix, urlPath, urlPathPre, urlPathSur,
			// )
			r.URL.Path = urlPathSur
			filecache.HttpHandler(cache)(w, r)
		},
	)

	srvAddr := fmt.Sprintf(":%d", *fPort)
	fmt.Printf("serving %s on %s\n", srvWD, srvAddr)
	if err := http.ListenAndServe(srvAddr, nil); err != nil {
		log.Fatal(err)
	}
}

func displayCacheStats(cache *filecache.FileCache) {
	fmt.Printf("-----[ cache stats: %s ]-----\n",
		time.Now().In(loc).Format(time.RFC3339))
	fmt.Printf("files cached: %d (max: %d)\n", cache.Size(),
		cache.MaxItems)
	fmt.Printf("cache size: %d bytes (will cache files up to %d bytes)\n",
		cache.FileSize(), cache.MaxSize)
	cachedFiles := cache.StoredFiles()
	fmt.Println("[ cached files ]")
	for _, name := range cachedFiles {
		fmt.Printf("\t* %s\n", name)
	}
	fmt.Printf("\n\n")
}
