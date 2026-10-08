# 博客第二版

这一版将文章原稿、模板和发布流程放在同一个可维护的项目里。网页后台支持富文本和 Markdown、预览、图片上传、分类、标签和草稿。公开页面由 Hugo 生成，访客阅读不需要登录。

## 当前交付状态

- 已导入当前归档中的 87 篇文章，保留原文章地址。
- 已从 Git 历史恢复 `vmstat` 和“音视频基础- 编解码 码控”的误删正文。
- “笔记 宏观经济”和“公司20周年”继续排除。
- 已实现首页、归档、分类、标签、全文搜索、明暗主题、移动端布局及 RSS。
- 已实现网页编辑器。`/admin/?demo=1` 仅在本机提供演示，不会写入仓库；正式 `/admin/` 使用 GitHub 登录。
- 已实现仅允许 `changan29` 登录的 OAuth 服务，带 PKCE、一次性登录状态和安全 Cookie；服务只监听本机端口。
- 已编写自动构建、校验、部署、上线版本检查和失败回退流程。
- **真实 OAuth 应用、服务器安装和上线仍需完成。准备好这些配置前，自动部署开关保持关闭。**

## 本机使用

使用 Hugo 0.167.0：

```bash
hugo server --bind 127.0.0.1
```

构建和验证：

```bash
hugo --minify
python3 scripts/validate_site.py
cd server
go test -race ./...
```

Hugo 输出到 `public/`，该目录不提交到原稿分支。编辑器的发布按钮保存原稿到 `blog-source` 分支，GitHub Actions 完成构建和发布。草稿标志开启时，文章不会出现在生成的网站中。GitHub 仓库目前是公开的，因此原稿和草稿提交后也能被仓库访客阅读；有私密草稿需求时，需改用私有内容仓库并调整登录配置。

## 正式接入步骤

1. 先备份服务器目录 `/data/oneyearago/changan29.github.io`、当前 Nginx 配置和相关服务配置。证书私钥只保存在服务器备份中，不上传 GitHub。
2. 为编辑器创建 GitHub OAuth App：主页 `https://www.oneyearago.me/`，回调 `https://www.oneyearago.me/cms-oauth/callback`。登录密钥放在服务器 `/etc/blog-oauth.env`，权限 `600`，不得提交到 Git。
3. 编译 `server/`，安装服务账号和 `blog-oauth.service`。将 Nginx 的 OAuth location 加入现有 HTTPS server，保留已更新的证书；检查配置后再 reload。
4. 将现有站点作为初始 release，创建 `/data/oneyearago/current` 链接指向它。服务器独立演示文件单独备份，并按需保存在 `legacy-extra/`；不要复制旧版生成的文章。
5. 安装 `server/deploy-site.sh` 为 root 所有的 `/usr/local/bin/blog-deploy`。建立独立的 `blog-deploy` 系统账号，安装 `blog-deploy.service` 和 `blog-deploy.timer`，由服务器每分钟读取公开的 `blog-site` 分支并部署；GitHub 无需持有服务器 SSH 私钥。
6. 完成备份、服务和 Nginx 验证后，在 GitHub 把 `BLOG_PUBLISH_ENABLED` variable 设为 `true`。工作流会保存生成结果并等待线上版本匹配，超时则报错。
7. 发布一篇临时测试文章，核对生成、部署、线上正文和图片；再删除测试文章并检查归档同步。真实流程通过后才算上线完成。

`master` 暂时保留旧生成站点；新原稿在 `blog-source`，新生成结果在 `blog-site`。GitHub Pages 的来源目前仍是旧分支，若希望同时更新，需要另外切换 Pages 来源。

## 回退

部署脚本为每次发布保留独立目录，原子切换 `current` 链接，健康检查失败会恢复上一个链接。需要整体退回旧站时，将 `current` 重新指向切换前备份目录；若撤回 Nginx 的接入改动，恢复服务器配置备份，执行 `nginx -t` 后 reload。

升级前的本地文件和 Git 历史备份在同级 `blog-backup-20261008/`；服务器备份必须在实际切换前补齐。本地备份不包含服务器私钥和额外文件。

## 迁移证据

`migration-report.json` 记录每篇原文章地址、标题、转换后的源文件、图片和代码块数量，以及是否从历史版本恢复。`scripts/validate_site.py` 对照检查生成结果，并阻止已删除文章重新出现。`toolchain.json` 记录锁定的工具版本和编辑器 bundle 校验值。

网页使用本地样式和原生 JavaScript，移除了旧版 jQuery、动画和字体运行依赖。文章中的既有外部图片地址仍保留，外部资源的可访问性不由本地生成保证。原站使用的 Giscus 评论在本版尚未接入。
