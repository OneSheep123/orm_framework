# ORM框架

## 使用
```go
go mod tidy #安装相关依赖
```

## 目录介绍
- orm: orm相关文件
- predicate.go: 用于组装查询语句条件
- select.go: 用于查询语句构建
- select_test.go: 测试查询语句
- types.go: 抽象化组装、执行、返回SQL的相关角色