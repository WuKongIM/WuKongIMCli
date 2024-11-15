
## 配置服务器信息（context）

```
wk context add demo --server http://127.0.0.1:5001 --token xxxx --description "WuKongIM Demo"
```

```

WuKongIM Configuration Context "demo"

  Description: WuKongIM Demo
  Server URLs: http://127.0.0.1:5001

```

## 服务启动和停止

```shell
# 启动
wk start

# 停止 
wk stop

# 重启
wk restart

# 检查
wk doctor

# 升级
wk upgrade
```


## 压力测试（bench）


###

```
wk bench --pub 10 --msgs 100000 --channelNum 10 --channelType 1
```

```
wk bench --pub 10 --msgs 100000 --channel  123,234,456
```

#### 一个人发送大量消息测试吞吐量

一个发布者向ch001发送10万条.

```
wk bench ch001 --pub 1 
```

```
23:33:51 Starting pub/sub benchmark [msgs=100,000, msgsize=16 B, pubs=1, subs=1]
23:33:51 Starting publisher, publishing 100,000 messages
Finished      0s [======================================================================================================================================================] 100%

Pub stats: 5,173,828 msgs/sec ~ 78.95 MB/sec
```

一个发布者向ch001发送1000万条 每条16B的消息

```
wk bench ch001 --s 1 --size 16 --msgs 10000000 
```

```
23:34:29 Starting pub/sub benchmark [msgs=10,000,000, msgsize=16 B, pubs=1, subs=1]
23:34:29 Starting publisher, publishing 10,000,000 messages
Finished      2s [======================================================================================================================================================] 100%

Pub stats: 4,919,947 msgs/sec ~ 75.07 MB/sec
```


#### 一个人发同时一个人收测试吞吐量

一个人发10万条消息一个人在线收10万条消息吞吐量测试

```
wk bench ch001 --pub 1 -sub 1 --size 16 
```

```
23:36:00 Starting pub/sub benchmark [msgs=100,000, msgsize=16 B, pubs=1, subs=1]
23:36:00 Starting subscriber, expecting 100,000 messages
23:36:00 Starting publisher, publishing 100,000 messages
Finished      0s [======================================================================================================================================================] 100%
Finished      0s [======================================================================================================================================================] 100%

WuKongIM Pub/Sub stats: 5,894,441 msgs/sec ~ 89.94 MB/sec
 Pub stats: 3,517,660 msgs/sec ~ 53.68 MB/sec
 Sub stats: 2,957,796 msgs/sec ~ 45.13 MB/sec

```

#### 一个人发N个人收测试消息吞吐量

```
wk bench ch001 --pub 1 --sub 5 --size 16 --msgs 1000000
```

```
23:38:08 Starting pub/sub benchmark [msgs=1,000,000, msgsize=16 B, pubs=1, subs=5]
23:38:08 Starting subscriber, expecting 1,000,000 messages
23:38:08 Starting subscriber, expecting 1,000,000 messages
23:38:08 Starting subscriber, expecting 1,000,000 messages
23:38:08 Starting subscriber, expecting 1,000,000 messages
23:38:08 Starting subscriber, expecting 1,000,000 messages
23:38:08 Starting publisher, publishing 1,000,000 messages
Finished      0s [======================================================================================================================================================] 100%
Finished      0s [======================================================================================================================================================] 100%
Finished      0s [======================================================================================================================================================] 100%
Finished      0s [======================================================================================================================================================] 100%
Finished      0s [======================================================================================================================================================] 100%
Finished      0s [======================================================================================================================================================] 100%

WuKongIM Pub/Sub stats: 7,123,965 msgs/sec ~ 108.70 MB/sec
 Pub stats: 1,188,419 msgs/sec ~ 18.13 MB/sec
 Sub stats: 5,937,525 msgs/sec ~ 90.60 MB/sec
  [1] 1,187,633 msgs/sec ~ 18.12 MB/sec (1000000 msgs)
  [2] 1,187,597 msgs/sec ~ 18.12 MB/sec (1000000 msgs)
  [3] 1,187,526 msgs/sec ~ 18.12 MB/sec (1000000 msgs)
  [4] 1,187,528 msgs/sec ~ 18.12 MB/sec (1000000 msgs)
  [5] 1,187,505 msgs/sec ~ 18.12 MB/sec (1000000 msgs)
  min 1,187,505 | avg 1,187,557 | max 1,187,633 | stddev 48 msgs
```

#### N个人发N个人收测试消息吞吐量

```
wk bench  --pub 5 --sub 5 --size 16 --msgs 1000000
```

```
23:39:28 Starting pub/sub benchmark [msgs=1,000,000, msgsize=16 B, pubs=5, subs=5]
23:39:28 Starting subscriber, expecting 1,000,000 messages
23:39:28 Starting subscriber, expecting 1,000,000 messages
23:39:28 Starting subscriber, expecting 1,000,000 messages
23:39:28 Starting subscriber, expecting 1,000,000 messages
23:39:28 Starting subscriber, expecting 1,000,000 messages
23:39:28 Starting publisher, publishing 200,000 messages
23:39:28 Starting publisher, publishing 200,000 messages
23:39:28 Starting publisher, publishing 200,000 messages
23:39:28 Starting publisher, publishing 200,000 messages
23:39:28 Starting publisher, publishing 200,000 messages
Finished      0s [======================================================================================================================================================] 100%
Finished      0s [======================================================================================================================================================] 100%
Finished      0s [======================================================================================================================================================] 100%
Finished      0s [======================================================================================================================================================] 100%
Finished      0s [======================================================================================================================================================] 100%
Finished      0s [======================================================================================================================================================] 100%
Finished      0s [======================================================================================================================================================] 100%
Finished      0s [======================================================================================================================================================] 100%
Finished      0s [======================================================================================================================================================] 100%
Finished      0s [======================================================================================================================================================] 100%

WuKongIM Pub/Sub stats: 7,019,849 msgs/sec ~ 107.11 MB/sec
 Pub stats: 1,172,667 msgs/sec ~ 17.89 MB/sec
  [1] 236,240 msgs/sec ~ 3.60 MB/sec (200000 msgs)
  [2] 236,168 msgs/sec ~ 3.60 MB/sec (200000 msgs)
  [3] 235,541 msgs/sec ~ 3.59 MB/sec (200000 msgs)
  [4] 234,911 msgs/sec ~ 3.58 MB/sec (200000 msgs)
  [5] 235,545 msgs/sec ~ 3.59 MB/sec (200000 msgs)
  min 234,911 | avg 235,681 | max 236,240 | stddev 485 msgs
 Sub stats: 5,851,064 msgs/sec ~ 89.28 MB/sec
  [1] 1,171,181 msgs/sec ~ 17.87 MB/sec (1000000 msgs)
  [2] 1,171,169 msgs/sec ~ 17.87 MB/sec (1000000 msgs)
  [3] 1,170,867 msgs/sec ~ 17.87 MB/sec (1000000 msgs)
  [4] 1,170,641 msgs/sec ~ 17.86 MB/sec (1000000 msgs)
  [5] 1,170,250 msgs/sec ~ 17.86 MB/sec (1000000 msgs)
  min 1,170,250 | avg 1,170,821 | max 1,171,181 | stddev 349 msgs
```

## 稳定性测试

#### 添加测试机器(用户模拟客户端连接)

```
wk  machine add IP:PORT IP:PORT ...
```

#### 上线用户

```
wk test online 10000 
```


## 命令行聊天器

#### 连接IM

```
wk connect  [uid] [token]
```

### 创建用户

创建前缀为usr的100个用户 (usr1,usr2,usr3.....)

```

wk user create --prefix usr --num 100

```

### 创建频道

创建前缀为ch的100个频道 (ch1,ch2,ch3.....)(默认频道类型为2 即群聊频道)

```
wk channel create --prefix ch --num 100
```

## 订阅者（subscriber）

### 添加订阅者

添加指定用户到指定前缀的频道

```

wk subscriber add --chPrefix ch --chType=2 --chNum=10 --list u1,u2,u3

```
每个频道前缀为ch的频道添加10000个前缀为usr的订阅者 （订阅者需要通过创建用户创建）

```
wk subscriber add --chPrefix ch --chType=2 --chNum=50 --subPrefix usr --subNum 10000
```

### 移除订阅者

移除指定用户到指定前缀的频道

```

wk subscriber remove --chPrefix ch --chType=2 --chNum=10 --list u1,u2,u3

```

移除每个频道前缀为ch的频道移除10000个前缀为usr的订阅者

```

wk subscriber remove --chPrefix ch --chType=2 --chNum=50 --subPrefix usr --subNum 10000

```

## 黑明单（denylist）

### 添加黑名单

添加指定用户到指定前缀的频道黑明单

```

wk denylist add --chPrefix ch --chType=2 --chNum=10 --list u1,u2,u3

```

每个频道前缀为ch的频道添加10000个前缀为usr的黑名单 

```

wk denylist add --chPrefix ch --chType=2 --chNum=50 --subPrefix usr --subNum 10000

```

### 移除黑名单

移除指定用户到指定前缀的频道黑名单

```

wk denylist remove --chPrefix ch --chType=2 --chNum=10 --list u1,u2,u3

```

移除每个频道前缀为ch的频道移除10000个前缀为usr的黑名单

```

wk denylist remove --chPrefix ch --chType=2 --chNum=50 --subPrefix usr --subNum 10000

```

## 白名单（allowlist）


### 添加白名单

添加指定用户到指定前缀的频道白名单

```

wk allowlist add --chPrefix ch --chType=2 --chNum=10 --list u1,u2,u3

```

每个频道前缀为ch的频道添加10000个前缀为usr的白名单 

```

wk allowlist add --chPrefix ch --chType=2 --chNum=50 --subPrefix usr --subNum 10000

```

### 移除白名单

移除指定用户到指定前缀的频道白名单

```
wk allowlist remove --chPrefix ch --chType=2 --chNum=10 --list u1,u2,u3

```

移除每个频道前缀为ch的频道移除10000个前缀为usr的白名单

```

wk allowlist remove --chPrefix ch --chType=2 --chNum=50 --subPrefix usr --subNum 10000

```


## mock命令

#### 模拟上线

```
wk mock online --num 15000

num: 在线用户数量
```

#### 模拟单聊

模拟1000个用户每隔5秒随机向一个频道发送一条消息（频道前缀为ch 频道数量为100个）

```
wk mock chat --num 1000 --prefix=usr --interval 5s --chPrefix ch --chNum 100
```

模拟1000个用户每隔5秒随机向一个用户发送一条消息

```
wk mock chat --num 1000 --prefix=usr --interval 5s
```


<!-- 
1000同时在线用户，每个用户每5秒发送一次消息

``` 
wk mock chat --num 1000 --prefix=usr --interval 5s

num: 同时用户在线数量
prefix: 用户id前缀
interval: 在线用户发送消息间隔时间 单位毫秒
```


#### 模拟群聊

模拟100个群 每个群500人 每个群成员在线率50% 每个群成员每5秒发送一次消息

```

wk mock chat --type group  --num 100 --prefix=ch  --subNum 500 --subPrefix=usr --onlineRate 0.5 --interval 5s

type: 模拟类型 默认为单聊
num: 群数量，模拟创建的群数量
prefix: 群id前缀
subNum: 每个群的成员数量
subPrefix: 订阅者前缀
onlineRate: 每个群的成员在线率
interval: 在线群成员发送消息间隔时间 单位毫秒

``` -->