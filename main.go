package main

import (
	"bytes"
	"fmt"
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

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigDefault

func main() {

	baseName := "dev"
	if len(os.Args) > 1 {
		baseName = os.Args[1]
	}
	fmt.Printf("os args:%#v", os.Args)
	if baseName != "" {
		//向哪个branch 发merge request
		fmt.Printf("os arg1: %#v\n", baseName)
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

	fmt.Printf("branches: %#v\n", pp(branches))
	fmt.Printf("head: %#v\n", pp(hp))
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
	fmt.Printf("commit Msg: %#v", commitMsg)

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
	fmt.Printf("remote url: %#v, len: %#v\n", rawRemoteUrl, len(rawRemoteUrl))
	//fmt.Printf("ssp url: %#v, len: %#v\n", ssp, len(ssp))
	//fmt.Printf("rUrl: [%#v]\n", strings.TrimSpace(ssp[1]))
	//return

	u, err := url.Parse(rawRemoteUrl)
	if err != nil {
		panic(err)
	}
	r14 := len(u.Path) - 4
	repoName := u.Path[1:r14]
	fmt.Printf("remote name: %#v\n", repoName)

	cbName := strings.ReplaceAll(string(hp.Name()), "refs/heads/", "")
	fmt.Printf("current branch name:%#v\n", cbName)

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
	fmt.Printf("ghUrl: %#v\n", ghurl)
	//bodyStr := fmt.Sprintf(`{"head":"%s","base":"%s", "title":"%s"}`, cbName, baseName, commitMsg)
	githubApiBody := map[string]string{
		"head":  cbName,
		"base":  baseName,
		"title": commitMsg,
	}
	fmt.Printf("githubApiBody: %#v\n", pp(githubApiBody))
	jsonStr, _ := json.Marshal(githubApiBody)
	postBody := bytes.NewBuffer(jsonStr)

	host := "github.com"
	err, response := makeHttpRequest(ghurl, postBody, host)
	if err != nil {
		panic(err)
	}

	fmt.Printf("response: %#v\n", response)
	bodyAll, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("api response: %#v,\n body text: %#v", pp(response), pp(string(bodyAll)))
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
	fmt.Printf("glUrl: %#v\n", ghurl)
	//bodyStr := fmt.Sprintf(`{"head":"%s","base":"%s", "title":"%s"}`, cbName, baseName, commitMsg)
	githubApiBody := map[string]string{
		"id":            eRepoName,
		"source_branch": cbName,
		"target_branch": baseName,
		"title":         commitMsg,
	}
	fmt.Printf("gitlabApiBody: %#v\n", pp(githubApiBody))
	jsonStr, _ := json.Marshal(githubApiBody)
	postBody := bytes.NewBuffer(jsonStr)

	err, response := makeHttpRequest(ghurl, postBody, gitHost)
	if err != nil {
		panic(err)
	}
	bodyAll, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("api response: %#v,\n body text: %#v\n", pp(response), pp(string(bodyAll)))
	if response.StatusCode == 200 || response.StatusCode == 201 {
		var bodyStructs map[string]interface{}
		err := json.Unmarshal(bodyAll, &bodyStructs)
		if err != nil {
			panic(err)
		}
		openUrlInBrowser(bodyStructs["web_url"].(string))
	}
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
			fmt.Printf(" error: %#v\n", err)
		} else {

			clt.Transport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
		}
	}
	response, err := clt.Do(hc)
	return err, response
}

//按照指定宽度，插入换行\n
func pp(v interface{}) string {

	fstr := fmt.Sprintf("%v", v)
	if len(fstr) > 80 {

		return prettyPrint(fstr, 80, "\n")
	}
	return fstr
}
func prettyPrint(body string, limit int, end string) string {

	var charSlice []rune

	// push characters to slice
	for _, char := range body {
		charSlice = append(charSlice, char)
	}

	var result string = ""

	for len(charSlice) >= 1 {
		// convert slice/array back to string
		// but insert end at specified limit
		if limit > len(body) {
			limit = len(body)
		}

		result = result + string(charSlice[:limit]) + end

		// discard the elements that were copied over to result
		charSlice = charSlice[limit:]

		// change the limit
		// to cater for the last few words in
		// charSlice
		if len(charSlice) < limit {
			limit = len(charSlice)
		}

	}

	return result

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

		fmt.Printf(" error: %#v", err.Error())
	}
}
