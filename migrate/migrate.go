package migrate

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/go-github/v52/github"
	"github.com/pkg/errors"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"
	"golang.org/x/oauth2"

	"github.com/marwan-at-work/mod"
	"github.com/marwan-at-work/mod/major"
)

// Run looks into the CWD go.mod file,
// clones all +incompatible dependencies
// and migrates them on your behalf to Go Modules.
func Run(githubToken string, limit int, test bool) error {
	f, err := mod.GetModFile(".")
	if err != nil {
		return errors.Wrap(err, "could not get mod file")
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubToken})
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	gc := github.NewClient(tc)
	urls := []string{}
	var count int
	pch := make(chan resp)
	for _, r := range f.Require {
		if strings.HasSuffix(r.Mod.Version, "+incompatible") {
			if limit > -1 && count > limit {
				break
			}
			count++
			go func(r *modfile.Require) {
				err := migrate(r.Mod.Path, gc, test)
				pch <- resp{err, "https://" + r.Mod.Path}
			}(r)
		}
	}

	for i := 0; i < count; i++ {
		res := <-pch
		if res.err != nil {
			fmt.Println("error migrating", res.url, ":", res.err)
			continue
		}
		urls = append(urls, res.url)
	}

	for _, p := range urls {
		exec.Command("open", p).Run()
	}

	return nil
}

type resp struct {
	err error
	url string
}

func migrate(path string, gc *github.Client, test bool) error {
	fmt.Printf("git clone %v\n", path)
	tempdir, err := ioutil.TempDir("", strings.Replace(path, "/", "_", -1))
	if err != nil {
		return errors.Wrap(err, "tempdir err")
	}
	defer os.RemoveAll(tempdir)
	red, err := deduceVanity(path)
	if err != nil {
		return errors.Wrap(err, "error deducing vanity")
	}
	dir, err := gitclone(red.gitURL, tempdir)
	if err != nil {
		return errors.Wrap(err, "error running git clone")
	}
	if checkMod(dir) {
		fmt.Printf("%v already migrated to modules\n", path)
		return nil
	}
	err = modinit(dir, path)
	if err != nil {
		return errors.Wrap(err, "err running go mod init")
	}
	fmt.Printf("downloading dependencies for %v\n", path)
	err = getdeps(dir)
	if err != nil {
		return errors.Wrap(err, "err running go get ./...")
	}
	t := getTag(dir)
	fmt.Printf("upgrading %v to v%v\n", path, t)
	if err := major.Run(dir, "upgrade", "", t); err != nil {
		return errors.Wrap(err, "err upgrading import paths")
	}

	err = rewriteGitIgnore(dir)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(path, "github.com/") {
		fmt.Printf("%v is updated, fork and contribute!\n", path)
		return nil
	}

	ss := strings.Split(path, "/")
	owner, repo := ss[1], ss[2]
	r, _, err := gc.Repositories.CreateFork(context.Background(), owner, repo, &github.RepositoryCreateForkOptions{})
	if _, ok := err.(*github.AcceptedError); !ok {
		return errors.Wrap(err, "error creating a fork")
	}
	time.Sleep(time.Second * 2)
	remote := r.GetCloneURL()
	err = addRemote(dir, remote)
	if err != nil {
		return errors.Wrap(err, "error adding remote")
	}

	err = checkout(dir)
	if err != nil {
		return errors.Wrap(err, "error checking out new branch")
	}

	err = add(dir)
	if err != nil {
		errors.Wrap(err, "error tracking files")
	}

	err = commit(dir)
	if err != nil {
		errors.Wrap(err, "error committing changes")
	}

	err = push(dir)
	if err != nil {
		errors.Wrap(err, "error pushing to github")
	}

	if test {
		fmt.Println(gc.Repositories.Delete(context.Background(), "marwan-at-work", repo))
	}

	return nil
}

func gitclone(url, tempdir string) (string, error) {
	c := exec.Command("git", "clone", url)
	var bts bytes.Buffer
	c.Stderr = &bts
	c.Dir = tempdir

	err := c.Run()
	if err != nil {
		return "", fmt.Errorf("output: %s", bts.String())
	}
	ff, err := ioutil.ReadDir(tempdir)
	if err != nil {
		return "", err
	}
	if len(ff) != 1 {
		return "", fmt.Errorf("expected one file but got %v", len(ff))
	}

	return filepath.Join(tempdir, ff[0].Name()), nil
}

func add(dir string) error {
	c := exec.Command("git", "add", "-A")
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Dir = dir

	return c.Run()
}

func push(dir string) error {
	c := exec.Command("git", "push", "automod", "automod")
	var bts bytes.Buffer
	c.Stderr = &bts
	c.Dir = dir

	err := c.Run()
	if err != nil {
		return fmt.Errorf("output: %s", bts.String())
	}

	return nil
}

func commit(dir string) error {
	c := exec.Command("git", "commit", "-m", "Migrate to Go Modules")
	var bts bytes.Buffer
	c.Stderr = &bts
	c.Dir = dir

	err := c.Run()
	if err != nil {
		return fmt.Errorf("output: %s", bts.String())
	}

	return nil
}

func checkout(dir string) error {
	c := exec.Command("git", "checkout", "-b", "automod")
	var bts bytes.Buffer
	c.Stderr = &bts
	c.Dir = dir

	err := c.Run()
	if err != nil {
		return fmt.Errorf("output: %s", bts.String())
	}

	return nil
}

func addRemote(dir, remote string) error {
	c := exec.Command("git", "remote", "add", "automod", remote)
	c.Stderr = os.Stderr
	c.Dir = dir

	return c.Run()
}

func checkMod(dir string) bool {
	f := filepath.Join(dir, "go.mod")
	_, err := os.Stat(f)
	return !os.IsNotExist(err)
}

func getTag(dir string) int {
	c := exec.Command("git", "tag", "-l", "v*")
	var b bytes.Buffer
	c.Stdout = &b
	c.Stderr = os.Stderr
	c.Dir = dir

	err := c.Run()
	if err != nil {
		panic(err)
	}
	tags := b.String()
	lines := strings.Split(tags, "\n")
	t := "v0.0.0"
	for _, l := range lines {
		if semver.IsValid(l) && semver.Compare(l, t) >= 1 {
			t = l
		}
	}
	m := semver.Major(t)
	if m == "" {
		return 0
	}
	numStr := m[1:]
	i, err := strconv.Atoi(numStr)
	if err != nil {
		panic(err)
	}

	return i
}

func modinit(dir, path string) error {
	c := exec.Command("go", "mod", "init", path)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Dir = dir
	c.Env = append(os.Environ(), "GO111MODULE=on")

	return c.Run()
}

func getdeps(dir string) error {
	c := exec.Command("go", "get", "./...")
	c.Dir = dir
	c.Env = append(os.Environ(), "GO111MODULE=on")

	return c.Run()
}

type redir struct {
	vcs    string
	base   string
	gitURL string
}

func deduceVanity(path string) (redir, error) {
	var r redir
	u, err := url.Parse(path)
	if err != nil {
		return r, err
	}
	u.Scheme = "http"
	u.RawQuery = "go-get=1"

	resp, err := http.Get(u.String())
	if err != nil {
		return r, err
	}
	defer resp.Body.Close()
	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return r, err
	}

	document.Find("meta").Each(func(i int, s *goquery.Selection) {
		val, _ := s.Attr("name")
		if val != "go-import" {
			return
		}

		cnt, _ := s.Attr("content")
		fields := strings.Fields(cnt)
		if len(fields) != 3 {
			return // return err
		}
		r.base = fields[0]
		r.vcs = fields[1]
		r.gitURL = fields[2]
	})

	return r, nil
}

func rewriteGitIgnore(dir string) error {
	p := filepath.Join(dir, ".gitignore")
	f, err := os.Open(p)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "could not open gitignore")
	}
	lines := []string{}
	scnr := bufio.NewScanner(f)
	for scnr.Scan() {
		line := scnr.Text()
		if line != "go.mod" && line != "go.sum" {
			lines = append(lines, line)
		}
	}
	if scnr.Err() != nil {
		f.Close()
		return errors.Wrap(err, "error while scanning")
	}
	f.Close()
	return ioutil.WriteFile(p, []byte(strings.Join(lines, "\n")+"\n"), 0o666)
}
