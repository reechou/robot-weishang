# maybe more powerful
# for mac (sed for linux is different)
dir=`echo ${PWD##*/}`
grep "robot-weishang" * -R | grep -v Godeps | awk -F: '{print $1}' | sort | uniq | xargs sed -i '' "s#robot-weishang#$dir#g"
mv robot-weishang.ini $dir.ini

