# gor-proxy

使用gor复制流量时，gor 1.3.0 好像不支持变更域名    

```bash
m.163.cn ===> m.sandbox.163.cn
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

