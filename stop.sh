#!/bin/bash
pwd=`pwd`
name=`pwd`/gameworld
kill -2 `ps axu | grep $name |grep -v grep| awk '{print $2}'`

echo "kill "$name" "
