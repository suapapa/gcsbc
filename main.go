package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gokyle/filecache"
)

func main() {
	fExpire := flag.Int("e",
		0, // filecache.DefaultExpireItem
		"maximum number of seconds between accesses a file can stay in the cache")
	fGarbage := flag.Int("g", filecache.DefaultEvery,
		"scan the cache for expired items every <n> seconds")
	fItems := flag.Int("n",
		100, //filecache.DefaultMaxItems,
		"max number of files to store in the cache")
	fPort := flag.Int("p", 8080, "port to listen on")
	fChroot := flag.Bool("r", false, "chroot to the working directory")
	fSize := flag.Int64("s", filecache.DefaultMaxSize, "max file size to cache")
	fUser := flag.String("u", "", "user to run as")
	fSubPathRoot := flag.String("f", "", "remove prefix subpath in url.PATH")
	fDumpCache := flag.String("d", "",
		"dump cache stats duration; by default, this is turned off. Must be parsable with time.ParseDuration.")
	flag.Parse()

	srvWD := "."
	if flag.NArg() > 0 {
		srvWD = flag.Arg(0)
	}
	if *fChroot {
		srvWD = chroot(srvWD)
	}

	if *fUser != "" {
		setuid(*fUser)
	}

	srvWD, err := filepath.Abs(srvWD)
	chk(err)
	err = os.Chdir(srvWD)
	chk(err)
	cache := &filecache.FileCache{
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

	srvAddr := fmt.Sprintf(":%d", *fPort)
	fmt.Printf("serving %s on %s\n", srvWD, srvAddr)
	http.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			if *fSubPathRoot != "" {
				origURLPath := strings.Trim(r.URL.Path, "/")
				origPaths := strings.Split(origURLPath, "/")
				if len(origPaths) > 0 && origPaths[0] == strings.Trim(*fSubPathRoot, "/") {
					r.URL.Path = "/" + strings.Join(origPaths[1:], "/")
				}
			}

			filecache.HttpHandler(cache)(w, r)
		},
	)
	if err := http.ListenAndServe(srvAddr, nil); err != nil {
		log.Fatal(err)
	}
}

func setuid(username string) {
	usr, err := user.Lookup(username)
	chk(err)
	uid, err := strconv.Atoi(usr.Uid)
	chk(err)
	err = syscall.Setreuid(uid, uid)
	chk(err)
}

func chroot(path string) string {
	err := syscall.Chroot(path)
	chk(err)
	return "/"
}

func displayCacheStats(cache *filecache.FileCache) {
	fmt.Printf("-----[ cache stats: %s ]-----\n",
		time.Now().Format("2006-01-02 15:04:05"))
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
