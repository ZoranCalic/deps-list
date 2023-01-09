cd ..
rm -rf _org_repos
mkdir -p _org_repos
cd _org_repos

git clone $1
cd $2

if [ -f "Gemfile" ]
then
    bundle list
fi
