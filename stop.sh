#!/bin/bash
pwd=`pwd`
name=`pwd`/gameworld
pid=`ps aux |grep $name|grep -v grep|grep -v "/bin/bash"`
echo "pid=${pid}"

echo "kill "$name""
