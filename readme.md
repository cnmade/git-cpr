# git-cpr

一个快速创建 pull request 的工具。

## 编译

为了让git 能跟git-cpr，也就是本软件协同工作，你需要构建一下 git-cpr

请准备Golang 1.17+以上版本。

```bash
go install github.com/cnmade/git-cpr
```

## 环境需求

1. linux/windows/macos
2. git
3. 复制git-cpr到git相同目录下
4. 确保~/.netrc或者~/_netrc里面配置了正确的账号密码
5. 确保git origin 是https协议
6. 确保提交pull 的仓库有dev，master分支
7. 确保网络通畅
8. 确保你己经建立了工作分支，如for/dev
9. 确保代码开发完，提交commit，并推送到origin
10. 执行 git cpr 可以快速创建一个merge request

## 支持的git 平台

1. github.com
2. gitlab.com
3. 自己托管的gitlab实例

## 使用

运行命令
```bash
git cpr [可选，如dev或者master]
```
表示建立新的pull request，目标合并branch 是dev或者是master

您也可以配合SmartGit这样的工具来使用。

![image](https://user-images.githubusercontent.com/278153/151919571-7c001d07-647a-4aa3-83ec-b92e7ad2dc67.png)

![image](https://user-images.githubusercontent.com/278153/151919593-47209160-883a-4c0a-ad16-748bd2fab9ad.png)

![image](https://user-images.githubusercontent.com/278153/151919611-8861ef42-5c92-4430-8a5f-98e51050dcd4.png)
