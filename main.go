package main

import (
	"bytes"
	"fmt"
	"github.com/cnmade/goencode/json"
	"github.com/cnmade/gonetrc"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func main() {

	json.UnsupportedBehaviour = json.UnsupportedBehaviourWithNull

	baseName := "dev"
	if len(os.Args) > 1 {
		baseName = os.Args[1]
	}
	fmt.Printf("os args:%#v \n", os.Args)
	if baseName != "" {
		//向哪个branch 发merge request
		fmt.Printf("os arg1: %s \n", baseName)
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

	fmt.Printf("branches: %s \n", pp(branches))
	fmt.Printf("head: %s \n", pp(hp))
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
	fmt.Printf("commit Msg: %s \n", commitMsg)

	rawCommitMsg := []rune(commitMsg)
	if len(rawCommitMsg) > 50 {
		commitMsg = string(rawCommitMsg[:50])
		fmt.Printf("new commitMsg after substr: %s \n", commitMsg)
	}

	remote, err := r.Remote("origin")
	if err != nil {
		panic(err)
	}
	rawUrl := strings.Fields(remote.String())

	if len(rawUrl) < 1 {
		panic("无法获取远程地址 \n")
	}
	rawRemoteUrl := rawUrl[1]
	//ssp := strings.Fields(rawRemoteUrl)
	fmt.Printf("remote url: %s, len: %v \n", rawRemoteUrl, len(rawRemoteUrl))
	//fmt.Printf("ssp url: %#v, len: %#v\n", ssp, len(ssp))
	//fmt.Printf("rUrl: [%#v]\n", strings.TrimSpace(ssp[1]))
	//return

	u, err := url.Parse(rawRemoteUrl)
	if err != nil {
		panic(err)
	}
	r14 := len(u.Path) - 4
	repoName := u.Path[1:r14]
	fmt.Printf("remote name: %s \n", repoName)

	cbName := strings.ReplaceAll(string(hp.Name()), "refs/heads/", "")
	fmt.Printf("current branch name:%s \n", cbName)

	switch u.Host {
	case "github.com":

		githubCreateNewPr(repoName, cbName, baseName, commitMsg)
	case "gitlab.com":
		gitlabCreateNewPr("gitlab.com", repoName, cbName, baseName, commitMsg)
	default:
		gitlabCreateNewPr(u.Host, repoName, cbName, baseName, commitMsg)

	}

}

func githubCreateNewPr(repoName, cbName, baseName, commitMsg string) {

	//Github
	ghurl := fmt.Sprintf("https://api.github.com/repos/%s/pulls", repoName)
	fmt.Printf("ghUrl: %s \n", ghurl)
	//bodyStr := fmt.Sprintf(`{"head":"%s","base":"%s", "title":"%s"}`, cbName, baseName, commitMsg)
	githubApiBody := map[string]string{
		"head":  cbName,
		"base":  baseName,
		"title": commitMsg,
	}
	fmt.Printf("githubApiBody: %s \n", pp(githubApiBody))
	jsonStr, _ := json.Marshal(githubApiBody)
	postBody := bytes.NewBuffer(jsonStr)

	host := "github.com"
	err, response := makeHttpRequest(ghurl, postBody, host)
	if err != nil {
		panic(err)
	}

	responseJsonStr, err := json.Marshal(response)
	if err != nil {
		ProcessError(err)
	}
	fmt.Printf("response: %s \n", pp(string(responseJsonStr)))
	bodyAll, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("api response body text: %s \n", pp(string(bodyAll)))
	if response.StatusCode == 200 || response.StatusCode == 201 {
		var bodyStructs map[string]interface{}
		err := json.Unmarshal(bodyAll, &bodyStructs)
		if err != nil {
			panic(err)
		}
		openUrlInBrowser(bodyStructs["html_url"].(string))
	}
}

func gitlabCreateNewPr(gitHost, repoName, cbName, baseName, commitMsg string) {

	//Github
	eRepoName := url.QueryEscape(repoName)
	ghurl := fmt.Sprintf("https://%s/api/v4/projects/%s/merge_requests", gitHost, eRepoName)
	fmt.Printf("glUrl: %#v \n", ghurl)
	//bodyStr := fmt.Sprintf(`{"head":"%s","base":"%s", "title":"%s"}`, cbName, baseName, commitMsg)
	githubApiBody := map[string]string{
		"id":            eRepoName,
		"source_branch": cbName,
		"target_branch": baseName,
		"title":         commitMsg,
	}
	fmt.Printf("gitlabApiBody: %s \n", pp(githubApiBody))
	jsonStr, err := json.Marshal(githubApiBody)
	if err != nil {
		ProcessError(err)
	}
	postBody := bytes.NewBuffer(jsonStr)

	err, response := makeHttpRequest(ghurl, postBody, gitHost)
	if err != nil {
		panic(err)
	}

	responseJsonStr, err := json.Marshal(response)
	if err != nil {
		ProcessError(err)
	}
	fmt.Printf("response: %s \n", pp(string(responseJsonStr)))

	bodyAll, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("api response body text: %s", pp(string(bodyAll)))
	if response.StatusCode == 200 || response.StatusCode == 201 {
		var bodyStructs map[string]interface{}
		err := json.Unmarshal(bodyAll, &bodyStructs)
		if err != nil {
			panic(err)
		}
		openUrlInBrowser(bodyStructs["web_url"].(string))
	}
}

func ProcessError(err error) {
	fmt.Printf("error: %#v \n", err)
}

func makeHttpRequest(ghurl string, postBody *bytes.Buffer, host string) (error, *http.Response) {
	proxy_config := os.Getenv("HTTP_PROXY")
	fmt.Printf("proxy config: %#v\n", proxy_config)
	hc, err := http.NewRequest("POST", ghurl, postBody)
	if err != nil {
		panic(err)
	}

	u, p := gonetrc.GetCredentials(host)
	hc.SetBasicAuth(u, p)
	hc.Header.Add("PRIVATE-TOKEN", p)

	hc.Header.Add("Content-Type", "application/json")
	clt := http.Client{}
	if proxy_config != "" {
		proxyUrl, err := url.Parse(proxy_config)
		if err != nil {
			fmt.Printf(" error: %#v \n", err)
		} else {

			clt.Transport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
		}
	}
	response, err := clt.Do(hc)
	return err, response
}

//按照指定宽度，插入换行\n
func pp(v interface{}) string {

	fstr := ""
	switch v.(type) {
	case string:
		fstr = v.(string)
	default:
		fstr = fmt.Sprintf("%v", v)
	}

	nfstr := []rune(fstr)
	ofstr := chunkpp(nfstr, 80)
	outStr := ""
	for _, v := range ofstr {
		outStr += string(v) + "\n"
	}

	return outStr
}

func chunkpp(items []rune, chunkSize int) (chunks [][]rune) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}

	return append(chunks, items)
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

		fmt.Printf(" error: %s \n", err.Error())
	}
}
