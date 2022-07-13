package main

/*
=== Утилита wget ===

Реализовать утилиту wget с возможностью скачивать сайты целиком

Программа должна проходить все тесты. Код должен проходить проверки go vet и golint.
*/

import (
	"bytes"
	"container/list"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/purell"
	"github.com/anaskhan96/soup"
)

func main() {
	opts := new(options)
	opts.client = &http.Client{
		Timeout: 900 * time.Second,
	}
	if err := do(os.Args, opts); err != nil {
		if err != ErrGotErrorSomewhere {
			fmt.Fprintln(os.Stderr, err)
		}

		os.Exit(1)
	}
}

type options struct {
	recursive          bool
	args               []string
	urls               []*url.URL
	scheme             string
	host               string
	compareFileContent bool
	client             *http.Client
}

func (opts *options) parseFlags(args []string) {
	flagset := flag.NewFlagSet(args[0], flag.ExitOnError)
	flagset.BoolVar(&opts.recursive, "r", false, "Turn on recursive retrieving.")
	flagset.Parse(args[1:])

	opts.args = flagset.Args()
}

var ErrMissingURL = errors.New("missing URL argument")

func (opts *options) validate() error {
	if len(opts.args) == 0 {
		return ErrMissingURL
	}

	return nil
}

func (opts *options) complete() error {
	if opts.recursive {
		opts.compareFileContent = true
	}

	for _, arg := range opts.args {
		u, err := url.Parse(arg)
		if err != nil {
			return err
		}

		if !u.IsAbs() {
			u.Scheme = "http"
		}

		flags := purell.FlagsSafe | purell.FlagRemoveDotSegments | purell.FlagRemoveDuplicateSlashes | purell.FlagSortQuery
		normalized := purell.NormalizeURL(u, flags)
		nu, err := url.Parse(normalized)
		if err != nil {
			return err
		}

		if nu.Path == "" {
			nu.Path = "/"
		}

		opts.urls = append(opts.urls, nu)
	}

	return nil
}

func do(args []string, opts *options) error {
	opts.parseFlags(args)

	if err := opts.validate(); err != nil {
		return err
	}

	if err := opts.complete(); err != nil {
		return err
	}

	if !opts.recursive {
		return doWget(opts)
	}

	var err error
	for _, u := range opts.urls {
		if sErr := doWgetSite(u, opts); sErr != nil {
			err = sErr
		}
	}

	return err
}

var ErrGotErrorSomewhere = errors.New("got error somewhere")

func doWget(opts *options) error {
	gotError := false
	dir := "./"

	for _, u := range opts.urls {
		res, err := download(u, opts)
		if err != nil {
			gotError = true
			log.Println(err)
			continue
		}

		name := getFilename(u)
		if err := save(dir, name, res, opts); err != nil {
			gotError = true
			log.Println(err)
			continue
		}
	}

	if gotError {
		return ErrGotErrorSomewhere
	}

	return nil
}

func doWgetSite(u *url.URL, opts *options) error {
	gotError := false
	opts.scheme = u.Scheme
	opts.host = u.Host
	baseDir := "./" + opts.host + "/"

	added := make(map[string]bool)
	added[u.RequestURI()] = true

	todo := list.New()
	todo.PushBack(u)

	for todo.Len() != 0 {
		curr := todo.Remove(todo.Front()).(*url.URL)
		res, err := download(curr, opts)
		if err != nil {
			gotError = true
			log.Println(err)
			continue
		}

		name := getFilename(curr)
		dir := baseDir + getDir(curr)
		if err := save(dir, name, res, opts); err != nil {
			gotError = true
			log.Println(err)
			continue
		}

		if !strings.HasPrefix(res.contentType, "text/html") {
			// log.Printf("content-type is not 'text/html', it is '%s'", res.contentType)
			continue
		}

		links := extractLinks(res, opts)
		for _, link := range links {
			if added[link.RequestURI()] {
				continue
			}

			added[link.RequestURI()] = true
			todo.PushBack(link)
			// log.Println("added", link.String())
		}

		// Делаем паузу, чтобы проще было наблюдать за работой программы во время разработки.
		// time.Sleep(1 * time.Second)
	}

	if gotError {
		return ErrGotErrorSomewhere
	}

	return nil
}

type result struct {
	url         *url.URL
	body        []byte
	contentType string
}

func download(u *url.URL, opts *options) (*result, error) {
	resp, err := opts.client.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		errMsg := fmt.Sprintf(`%s "%s": %s`, resp.Request.Method, resp.Request.URL, resp.Status)
		return nil, errors.New(errMsg)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Println("downloaded", resp.Request.URL)

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(body)
	}

	return &result{u, body, contentType}, nil
}

func getFilename(u *url.URL) string {
	if strings.HasSuffix(u.Path, "/") {
		if u.RawQuery == "" {
			return "index.html"
		}

		return "index.html?" + u.RawQuery
	}

	parts := strings.Split(u.Path, "/")
	last := parts[len(parts)-1]
	if u.RawQuery == "" {
		return last
	}

	return last + "?" + u.RawQuery
}

func getDir(u *url.URL) string {
	if u.Path == "/" {
		return ""
	}

	relativePath := strings.Replace(u.Path, "/", "", 1)

	if strings.HasSuffix(relativePath, "/") {
		return relativePath
	}

	parts := strings.Split(relativePath, "/")
	parts[len(parts)-1] = ""
	return strings.Join(parts, "/")
}

func save(dir, name string, r *result, opts *options) error {
	err := os.MkdirAll(dir, 0775)
	if err != nil {
		return err
	}

	newName := name
	for i := 1; ; i++ {
		if _, err := os.Stat(dir + newName); err != nil {
			// Файла с таким именем не существует. Можно сохранять.
			break
		}

		// Файл с таким именем уже существует.
		if opts.compareFileContent {
			// Сверяем содержимое файлов.
			existedData, err := os.ReadFile(dir + newName)
			if err == nil {
				if bytes.Equal(existedData, r.body) {
					log.Println("equal files, skipped saving to", dir+newName)
					return nil
				}
			}
		}

		// Добавляем число к имени.
		newName = name + fmt.Sprintf(".%d", i)
	}

	if err := os.WriteFile(dir+newName, r.body, 0664); err != nil {
		return err
	}
	log.Println("saved to", dir+newName)

	return nil
}

func extractLinks(r *result, opts *options) []*url.URL {
	res := make([]*url.URL, 0)
	doc := soup.HTMLParse(string(r.body))
	links := doc.FindAll("a")

	for _, link := range links {
		href := link.Attrs()["href"]
		u, err := url.Parse(href)
		if err != nil {
			log.Println(err)
			continue
		}

		normalized, err := url.Parse(normalizeURL(u, r, opts))
		if err != nil {
			log.Println(err)
			continue
		}

		if normalized.Host != opts.host {
			continue
		}

		res = append(res, normalized)
	}

	return res
}

func normalizeURL(u *url.URL, source *result, opts *options) string {
	flags := purell.FlagsSafe | purell.FlagRemoveDotSegments | purell.FlagRemoveDuplicateSlashes | purell.FlagSortQuery

	if u.IsAbs() {
		if u.Path == "" {
			u.Path = "/"
		}

		return purell.NormalizeURL(u, flags)
	}

	u.Scheme = opts.scheme
	u.Host = opts.host

	if strings.HasPrefix(u.Path, "/") {
		return purell.NormalizeURL(u, flags)
	}

	u.Path = getBasePath(source.url) + u.Path
	return purell.NormalizeURL(u, flags)
}

func getBasePath(u *url.URL) string {
	if strings.HasSuffix(u.Path, "/") {
		return u.Path
	}

	parts := strings.Split(u.Path, "/")
	parts[len(parts)-1] = ""
	return strings.Join(parts, "/")
}
