#!/bin/bash

# 项目地址，/go 在 GOPATH 里面
baseProjectDir="/usr/local/gopath/src/github.com/sencydai/gameworld"

pwd=`pwd`

# targetFile 编译后的二进制文件
targetFile="${pwd}/gameworld-new"

# branch 编译的分支
branch="master"

# buildPkg 编译的包名，main.go 所在的包
buildPkg="github.com/sencydai/gameworld"

# buildResult 编译结果
buildResult=""

gitPull() {
  pushd .

  cd "$baseProjectDir"
  git checkout "$branch"
  git pull

  popd
}

goBuild() {
    buildResult=`go build -o "$targetFile" "$buildPkg" 2>&1`

    if [ -z "$buildResult" ]; then
      buildResult="success"
    fi
}

gitPull
goBuild

if [ "$buildResult" = "success" ]; then
  chmod +x $targetFile
else
  echo "build error $buildResult"
  exit
fi

echo "All Complete"