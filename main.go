package main

import (
	"bytes"
	"fmt"
	"github.com/cnmade/gonetrc"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"net/http"
	"os"
	"strings"
)

func main() {

	baseName := os.Args[1]
	if baseName != "" {
		//向哪个branch 发merge request
		fmt.Printf("os arg1: %+v\n", baseName)
	} else {
		//默认是dev
		baseName = "dev"
	}
	fs := osfs.New(".")
	if _, err := fs.Stat(git.GitDirName); err == nil {
		fs, err = fs.Chroot(git.GitDirName)
		if err != nil {
			panic(err)
		}
	}

	s := filesystem.NewStorageWithOptions(fs, cache.NewObjectLRUDefault(), filesystem.Options{KeepDescriptors: true})
	r, err := git.Open(s, fs)
	if err != nil {
		panic(err)
	}
	hp, err := r.Head()
	if err != nil {
		panic(err)
	}
	branches, err := r.Branches()
	if err != nil {
		panic(err)
	}

	fmt.Printf("branches: %+v\n", branches)
	fmt.Printf("head: %+v\n", hp)
	fmt.Println("current branch name: " + hp.Name() + " \n")

	remote, err := r.Remote("origin")
	if err != nil {
		panic(err)
	}

	rawRemoteUrl := strings.TrimSpace(remote.String())
	fmt.Printf("remote url: %+v, len: %+v\n", rawRemoteUrl, len(rawRemoteUrl))

	repoTheDomain := rawRemoteUrl[7:26]
	fmt.Printf("remote host: %+v\n", repoTheDomain)

	r14 := len(rawRemoteUrl) - 64
	repoName := rawRemoteUrl[26:r14]
	fmt.Printf("remote name: %+v\n", repoName)

	cbName := strings.ReplaceAll(string(hp.Name()), "refs/heads/", "")
	fmt.Printf("current branch name:%+v\n", cbName)

	if repoTheDomain == "https://github.com/" {
		//Github
		ghurl := fmt.Sprintf("https://api.github.com/repos/%s/pulls", repoName)
		fmt.Printf("ghUrl: %+v\n", ghurl)
		bodyStr := fmt.Sprintf(`{"head":"%s","base":"%s"}`, cbName, baseName)
		var jsonStr = []byte(bodyStr)
		postBody := bytes.NewBuffer(jsonStr)
		hc, err := http.NewRequest("POST", ghurl, postBody)
		if err != nil {
			panic(err)
		}

		u, p := gonetrc.GetCredentials("github.com")
		hc.SetBasicAuth(u, p)

		clt := http.Client{}
		response, err := clt.Do(hc)
		if err != nil {
			panic(err)
		}
		fmt.Printf("api response: %+v", response)

	} else {

		//GITLAB
	}

}
