# git-cpr

一个快速创建 pull request 的工具。

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