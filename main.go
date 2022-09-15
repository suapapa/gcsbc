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
	loc = time.FixedZone("KST", +9*60*60)
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
	fSubPathRoot := flag.String("f", "", "remove prefix subpath in url.PATH")
	fDumpCache := flag.String("d", "",
		"dump cache stats duration; by default, this is turned off. Must be parsable with time.ParseDuration.")
	flag.Parse()

	srvWD := "."
	if flag.NArg() > 0 {
		srvWD = flag.Arg(0)
	}

	srvWD, err := filepath.Abs(srvWD)
	chk(err)
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
			if *fSubPathRoot != "" {
				origURLPath := strings.Trim(r.URL.Path, "/")
				origPaths := strings.Split(origURLPath, "/")
				if len(origPaths) > 0 && origPaths[0] == strings.Trim(*fSubPathRoot, "/") {
					r.URL.Path = "/" + strings.Join(origPaths[1:], "/")
				}
			}
			// log.Printf("URL.Host=%s, URL.Path=%s", r.URL.Host, r.URL.Path)
			// q, e := url.QueryUnescape(r.URL.String())
			// log.Println(q, e)

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
		time.Now().In(loc).Format("2006-01-02 15:04:05"))
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

func chk(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
