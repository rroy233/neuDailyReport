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
  "password_encoded": false,
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
  "password_encoded": false,
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

#### 说明

terminate_wait_time: 程序延迟结束的时间(s)

password_encoded: 配置文件中的**学生密码**是否经过[base64编码](https://tool.oschina.net/encrypt?type=3)



### 运行使用

> 由于东B限制，外网IP可能无法正常完成上报

Mac or Linux

```shell
./文件名称
```

Windows

```
双击可执行文件
```

### 参数说明

```
-a 执行【体温上报-早】
-b 执行【体温上报-午】
-c 执行【体温上报-晚】
-d 执行【健康上报】
-t 仅验证学生账号可用性
-help 帮助

（如执行时不输入参数，则常驻内存运行自动按计划上报）
```

例如：

1. 立即执行【体温上报-早】+【健康上报】

   ```shell
   ./neuDailyReport -a -d
   或
   ./neuDailyReport.exe -a -d
   ```

2. 立即执行【体温上报-午】

   ```shell
   ./neuDailyReport -b
   或
   ./neuDailyReport.exe -b
   ```

3. 常驻内存运行

   ```
   ./neuDailyReport
   ```
   或使用自动后台运行脚本(unix)

   ```
   编译运行：
   bash buildrun.sh
   或运行：
   bash run.sh
   ```
   

## 自行编译

### 环境依赖

* go 1.17

### 编译

```shell
go build -o neuDailyReport
```

或

```shell
# 交叉编译
make
```

## License

MIT License.



