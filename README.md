# How to use?
* 将 `workers/index.ts` 中的内容部署到 cloudflare workers 上
* 添加自己的域名至 workers，并配置该域允许通过 grpc 流量
* （可选）为 workers 设置环境变量 `Authorization=114514:1919810`
* 编辑 `main.go` 中的 `ocRawURL` 为自己的域名
* 运行 `go run .` 在 `127.0.0.1:8080` 上启动 HTTP 代理