#!/usr/bin/env bash

# 뭔지 모르겠네
set -euo pipefail

read -rp "GitHub Username: " user
read -rp "Projectname: " projectname

git clone git@github.com:Gizzahub/gzh-manager-go.git "$projectname"
cd "$projectname"
rm -rf .git
find . -type f -exec sed -i "s/gzh-manager-go/$projectname/g" {} +
find . -type f -exec sed -i "s/[Gg]izzahub/$user/g" {} +
git init
git add .
git commit -m "initial commit"
git remote add origin "git@github.com:$user/$projectname.git"

echo "template successfully installed."

go run main.go

exit 0
