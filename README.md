# ORM框架

## 使用
```go
go mod tidy # 安装相关依赖
```

## 目录介绍

```
│   go.mod 
│   go.sum
│   main.go
│   README.md
├───orm
│   │   aggregate.go # 用于构建聚合内容
│   │   assignment.go # 含有Assignable标记接口，实现该接口用于赋值语句
│   │   builder.go # 封装一些轻量级的操作，持有一些公共字段
│   │   column.go # 用于构建对应列元素
│   │   db.go # db包含了方言、注册中心
│   │   dialect.go # 方言
│   │   error.go  # 暴露出去的错误
│   │   expression.go # 原生SQL
│   │   insert.go # insert语句构建
│   │   insert_test.go # insert单元测试
│   │   predicate.go # 构建断言
│   │   result.go # update/insert/delete结果集
│   │   select.go # select语句构建
│   │   select_test.go # select单元测试
│   │   types.go # 语句构建规范
│   │   
│   ├───internal
│   │   ├───errs
│   │   │       error.go
│   │   │       
│   │   └───valuer # unsafe或者反射对结果集进行处理
│   │           reflect.go
│   │           reflect_test.go
│   │           unsafe.go
│   │           unsafe_test.go
│   │           value.go
│   │           value_test.go
│   │           
│   └───model # 元数据以及注册中心
│           model.go
│           registry.go
│           registry_test.go
│           
├───reflect # 反射测试
│   │   fields.go
│   │   fields_test.go
│   │   func_call.go
│   │   func_call_test.go
│   │   iterate.go
│   │   iterate_test.go
│   │   
│   └───types
│           user.go
│           
└───unsafe # unsafe测试
        accessor.go
        accessor_test.go
```

