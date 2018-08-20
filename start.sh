#!/bin/bash
pwd=`pwd`
target=`pwd`/gameworld

if [ -f "${target}-new" ]; then
  echo "upgrading..."
  if [ -f "${target}-backup" ]; then
    backupdt=`date +%Y%m%d-%H`
	mv "${target}-backup" "${target}-backup-${backupdt}"
  fi
  
  mv ${target} ${target}-backup
  mv ${target}-new ${target}
  
  echo "upgrade Complete"
  sleep 3
fi

$target &
