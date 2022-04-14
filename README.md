# gor-proxy

使用gor复制流量时，gor 1.3.0 好像不支持变更域名    

```bash
m.akin.cn ===> m.sandbox.akin.cn
www.akin.cn ===> www.sandbox.akin.cn
tongji.akin.cn pass
m.akin.cn/common/api/weixinjssdk pass
m.sandbox.akin.com pass
```

运行    

```bash
go clone https://github.com/akin520/gor-proxy.git
cd gor-proxy
go build
./gor-proxy -port=8018 -proxy="http://192.168.2.170:80"
```

gor进行流量复制    

```bash
gor --input-file :80 --http-original-host --output-http="http://192.168.2.200:8018|20%"
```

配置说明

```yaml
prefix: "sandbox"          #域名变更
host:                      #匹配域名
  - "akin.cn"
  - "akin.com"
exhost:                    #排除域名，只要包含就行
  - "tongji"
  - "sandbox"
  - "static"
path:                      #排除的URI路径，只要包含就行
  - "weixinjssdk"
```

