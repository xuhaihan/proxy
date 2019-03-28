# proxy代理
  在配置文件配置免费代理配置源，通过api接口获取代理
## 环境要求
1.go版本至少1.9

2.拉取项目之前先安装需要依赖,使用以下命令：
```
   go get -u github.com/AceDarkknight/GoProxyCollector
```
## 运行步骤
  该项目包含编译完成的静态二进制文件，可以直接后台运行该二进制文件，也可以重新运行生成覆盖之前旧的二进制文件
  
1.执行启动命令
```
  nohup ./proxy >& /tmp/proxy.log &
```
  上面命令是后台运行二进制文件，将控制台输出重定向到/tmp/proxy.log 中，重定向路径自定义创建
  
2.查看是否运行成功
```
    netstat -nlp
```
    查看是否在8090端口启动，若有则后台启动成功
## 测试
1.获取代理
```
GET http://localhost:8090/get
```
```
{
ip: "123.121.105.180",
port: 8060,
location: "高匿_北京市海淀区联通",
source: "http://www.ip3366.net/?stype=1&page=4",
speed: 2
}

```

2.删除代理
```
GET http://localhost:8090/delete?ip=1.2.3.4
```

##  代理数据源
- http://www.xicidaili.com
- http://www.89ip.cn
- http://www.kxdaili.com/
- https://www.kuaidaili.com

## 配置其他代理
   如果想获取其他来源的代理，在collectorConfig.xml中进行配置
