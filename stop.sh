#!/bin/bash
pwd=`pwd`
name=`pwd`/gameworld
pid=`ps aux |grep $name|grep -v grep|grep -v "/bin/bash"| awk '{print $2}'`
echo "pid=${pid}"

if [-n "$pid"];then
  kill -s 2 $pid
  
  pid=`ps aux |grep $name|grep -v grep|grep -v "/bin/bash"| awk '{print $2}'`
  while [-n "$pid"];do
	sleep 1
  done
fi

echo "kill "$name""
