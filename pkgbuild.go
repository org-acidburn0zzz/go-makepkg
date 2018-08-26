package main

import "text/template"

var pkgbuildTemplate = template.Must(
	template.New("pkgbuild").Parse(
		`{{if ne .Maintainer ""}}# Maintainer: {{.Maintainer}}
{{end}}pkgname={{.PkgName}}
_pkgname={{.ProgramName}}
pkgver=${PKGVER:-autogenerated}
pkgrel={{if eq .PkgRel "1"}}${PKGREL:-1}{{else}}{{.PkgRel}}{{end}}
pkgdesc="{{.PkgDesc}}"
arch=('i686' 'x86_64')
license=('{{.License}}')
depends=({{range .Dependencies}}
	'{{.}}'{{end}}
)
makedepends=(
	'go'
	'git'{{range .MakeDependencies}}
	'{{.}}'{{end}}
)

source=(
	"$_pkgname::{{.RepoURL}}#branch=${BRANCH:-master}"{{range .Files}}
	"{{.Name}}"{{end}}
)

md5sums=(
	'SKIP'{{range .Files}}
	'{{.Hash}}'{{end}}
)

backup=({{range .Backup}}
	"{{.}}"{{end}}
)

pkgver() {
	if [[ "$PKGVER" ]]; then
		echo "$PKGVER"
		return
	fi

	cd "$srcdir/$_pkgname"
	local date=$(git log -1 --format="%cd" --date=short | sed s/-//g)
	local count=$(git rev-list --count HEAD)
	local commit=$(git rev-parse --short HEAD)
	echo "$date.${count}_$commit"
}

build() {
	cd "$srcdir/$_pkgname"

	if [ -L "$srcdir/$_pkgname" ]; then
		rm "$srcdir/$_pkgname" -rf
		mv "$srcdir/go/src/$_pkgname/" "$srcdir/$_pkgname"
	fi

	rm -rf "$srcdir/go/src"

	mkdir -p "$srcdir/go/src"

	export GOPATH="$srcdir/go"

	mv "$srcdir/$_pkgname" "$srcdir/go/src/"

	cd "$srcdir/go/src/$_pkgname/"
	ln -sf "$srcdir/go/src/$_pkgname/" "$srcdir/$_pkgname"

	echo ":: Updating git submodules"
	git submodule update --init

	echo ":: Building binary"
	go get -v \
		-gcflags "-trimpath $GOPATH/src"{{if ne .VersionVarName ""}} \
		-ldflags="-X main.{{.VersionVarName}}=$pkgver-$pkgrel"{{end}}{{if .IsWildcardBuild}} \
		./...{{end}}
}

package() {
	find "$srcdir/go/bin/" -type f -executable | while read filename; do
		install -DT "$filename" "$pkgdir/usr/bin/$(basename $filename)"
	done{{range .Files}}
	install -DT -m0755 "$srcdir/{{.Name}}" "$pkgdir/{{.Path}}"{{end}}
}
`))
