# 绿色施工中间件

## 模块描述
1. comm模块，用于沟通主服务器和中断设备
2. config模块，用于程序的配置
3. devices模块，实现和具体设备的通信
4. utilities模块，一些通用功能的实现
5. main.go 主程序


## 消息说明

1. 中间件和服务器通信格式  MsgType:Serv
如:{"MsgType":"Serv","Action":"DevList","Data":["500001",600001]}  //5 6区分不同类型
Action:
    - DevList 

2. 服务器通过中间件和设备通信 MsgType:Devices
如：{"MsgType":"Devices","ID":"500001","Action":"每日电量"}  //服务器发给中间件

{"MsgType":"Devices","ID":"500001","Action":"每日电量","Data":"600度"}   //中间件发给服务器
