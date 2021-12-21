package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/cnmade/gonetrc"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func main() {

	baseName := "dev"
	if len(os.Args) > 1 {
		baseName = os.Args[1]
	}
	fmt.Printf("os args:%+v", os.Args)
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

	commitIter, err := r.Log(&git.LogOptions{From: hp.Hash()})

	if err != nil {
		panic(err)
	}
	nextCommit, err := commitIter.Next()
	if err != nil {
		panic(err)
	}
	commitMsg := nextCommit.Message
	if err != nil {
		panic(err)
	}
	fmt.Printf("commit Msg: %+v", commitMsg)

	remote, err := r.Remote("origin")
	if err != nil {
		panic(err)
	}
	rawUrl := strings.Fields(remote.String())

	if len(rawUrl) < 1 {
		panic("无法获取远程地址")
	}
	rawRemoteUrl := rawUrl[1]
	//ssp := strings.Fields(rawRemoteUrl)
	fmt.Printf("remote url: %+v, len: %+v\n", rawRemoteUrl, len(rawRemoteUrl))
	//fmt.Printf("ssp url: %+v, len: %+v\n", ssp, len(ssp))
	//fmt.Printf("rUrl: [%+v]\n", strings.TrimSpace(ssp[1]))
	//return

	repoTheDomain := rawRemoteUrl[0:19]
	fmt.Printf("remote host: %+v\n", repoTheDomain)

	r14 := len(rawRemoteUrl) - 4
	repoName := rawRemoteUrl[19:r14]
	fmt.Printf("remote name: %+v\n", repoName)

	cbName := strings.ReplaceAll(string(hp.Name()), "refs/heads/", "")
	fmt.Printf("current branch name:%+v\n", cbName)

	if repoTheDomain == "https://github.com/" {
		//Github
		ghurl := fmt.Sprintf("https://api.github.com/repos/%s/pulls", repoName)
		fmt.Printf("ghUrl: %+v\n", ghurl)
		//bodyStr := fmt.Sprintf(`{"head":"%s","base":"%s", "title":"%s"}`, cbName, baseName, commitMsg)
		githubApiBody := map[string]string{
			"head":  cbName,
			"base":  baseName,
			"title": commitMsg,
		}
		fmt.Printf("githubApiBody: %+v\n", githubApiBody)
		jsonStr, _ := json.Marshal(githubApiBody)
		fmt.Printf("jsonStr: %+v\n", string(jsonStr))
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
		bodyAll, err := ioutil.ReadAll(response.Body)
		if err != nil {
			panic(err)
		}
		fmt.Printf("api response: %+v,\n body text: %+v", response, string(bodyAll))
		if response.StatusCode == 200 || response.StatusCode == 201 {
			var bodyStructs map[string]interface{}
			err := json.Unmarshal(bodyAll, &bodyStructs)
			if err != nil {
				panic(err)
			}
			openUrlInBrowser(bodyStructs["html_url"].(string))
		}

	} else {

		//GITLAB
	}

}

func openUrlInBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {

		fmt.Printf(" error: %+v", err.Error())
	}
}
