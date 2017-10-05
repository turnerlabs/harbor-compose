#!/bin/sh
set -e

app=harbor-compose

version=$(curl -s https://api.github.com/repos/turnerlabs/$app/releases/latest | grep 'tag_name' | cut -d\" -f4)

uname=$(uname)
suffix=""
case $uname in
"Darwin")
suffix="darwin_amd64"
;;
"Linux")
suffix="linux_amd64"
esac

url=https://github.com/turnerlabs/$app/releases/download/$version/ncd_$suffix
echo "Getting package $url"

curl -sSLo /usr/local/bin/$app $url
chmod +x /usr/local/bin/$app

echo "$app installed to /usr/local/bin/$app"