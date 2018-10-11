#/usr/bin/env bash
CONFIGURATIONS=(
    linux,386 \
    darwin,386 \
)

version=`cat VERSION`

for config in ${CONFIGURATIONS[@]}; do 
IFS=","
set $config

os="$1"
arch="$2"
base_path="build/$os/$arch"
path="$base_path"
mkdir -p "$path"
bin_name="pkgo"

if [ $os = "windows" ]; then
    bin_name="$bin_name.exe"
fi

GOOS="$os" GOARCH="$arch" go build -ldflags "-X main.Version=${version}" -o "$path/$bin_name"
tar czvf "build/pkgo-$version-$os-$arch.tar.gz" -C "$base_path" ${bin_name}

unset IFS;
done