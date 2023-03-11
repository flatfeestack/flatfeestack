#!/usr/bin/env sh

# https://github.com/ethers-io/ethers.js/issues/413

for FILE in $(ls *.sol)
do

  echo "$FILE\n"

  # print out struct tuple first
  awk '!/0$/{printf $0}/0$/' $FILE | sed 's/^.*\(struct[^}]*}\).*$/\1/g' | sed 's/struct[^{]*{ /tuple(/' | sed 's/ }/)/' | sed 's/;/,/g' | sed 's/, )/)/' | tr -s " " | sed 's/e( /e(/'

  echo "\n"

  # now turn events and functions into ethers.js human readable ABI form
  NAME=`echo $FILE | sed 's/\.sol//' | sed 's/^./\L&/'`
  echo "static ${NAME}ABI = ["
  egrep "event|function" $FILE | sed 's/^[ ]*\([e|f]\)/\1/' | sed s'/^\(.*\);$/\t"\1",/'
  echo "]\n"

done