
##  提交规范

仓库使用 [Conventional Commits](https://www.conventionalcommits.org/)：
`type(scope): 英文祈使句，小写开头，不加句号`。从 `git log` 归纳的实际用法：

- **type**：`feat` / `fix` / `docs` / `test` / `refactor` / `chore` / `ci` / `build`
- **scope**：模块目录名，如 `wecom` `wxkf` `wecomws` `tenant` `storage` `config` `k8s` `db`

一个 commit 只做一件事；文档和代码的变更可以同 commit，但不要把无关改动捎进来。

