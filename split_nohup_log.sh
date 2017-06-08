#!/bin/sh
#分割nohup产生的大文件
this_path=$(cd `dirname $0`;pwd)

cd $this_path
echo $this_path
current_date=`date -d "-1 day" "+%Y%m%d"`
echo $current_date
#一兆大小分割
split -b 1048576 -d -a 4 nohup.out   nohup_log/log_${current_date}_

cat /dev/null > nohup.out