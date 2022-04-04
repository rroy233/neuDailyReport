# 东北大学每日健康上报

> 一键完成健康上报（早中晚体温上报+健康上报）

## 使用

### 下载可执行文件

前往[Releases](https://github.com/rroy233/neuDailyReport/releases)下载可执行文件

### 编辑配置文件

在可执行文件同目录下，新建`config.json`

```json
{
  "terminate_wait_time": 3,
  "student_list": [{
    "stu_id": "统一身份认证学号",
    "password": "统一身份认证密码"
  }]
}
```

如果有多个账号可以这样写

```json
{
  "terminate_wait_time": 3,
	"student_list": [{
		"stu_id": "学号1",
		"password": "密码1"
	}, {
		"stu_id": "学号2",
		"password": "密码2"
	}, {
		"stu_id": "学号3",
		"password": "密码3"
	}]
}
```

保存文件

### 运行使用

Mac or Linux

```shell
./文件名称
```

Windows

```
双击可执行文件
```



## 自行编译

### 环境依赖

* go 1.17

### 编译

```shell
go build
```

或

```shell
# 交叉编译
make
```

## License

MIT License.



