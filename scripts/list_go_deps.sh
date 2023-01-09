cd ..
rm -rf _org_repos
mkdir -p _org_repos
cd _org_repos

git clone $1
cd $2

if [ -f "go.mod" ]
then
    go list -m -f '{{if not (or .Indirect .Main)}}{{.Path}} {{.Version}}{{end}}' all
fi
