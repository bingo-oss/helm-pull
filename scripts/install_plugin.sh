#!/bin/sh -e

if [ -n "${HELM_PUSH_PLUGIN_NO_INSTALL_HOOK}" ]; then
    echo "Development mode: not downloading versioned release."
    exit 0
fi

version="$(cat plugin.yaml | grep "version" | cut -d '"' -f 2)"
echo "Downloading and installing helm-pull v${version} ..."

url=""
if [ "$(uname)" = "Darwin" ]; then
    url="https://github.com/gzericlee/helm-pull/releases/download/${version}/helm-pull_${version}_darwin_amd64.tgz"
elif [ "$(uname)" = "Linux" ] ; then
    url="https://github.com/gzericlee/helm-pull/releases/download/${version}/helm-pull_${version}_linux_amd64.tgz"
else
    url="https://github.com/gzericlee/helm-pull/releases/download/${version}/helm-pull_${version}_windows_amd64.tgz"
fi

echo $url

mkdir -p "bin"
mkdir -p "releases/v${version}"

# Download with curl if possible.
if [ -x "$(which curl 2>/dev/null)" ]; then
    curl -sSL "${url}" -o "releases/v${version}.tar.gz"
else
    wget -q "${url}" -O "releases/v${version}.tar.gz"
fi
tar xzf "releases/v${version}.tar.gz" -C "releases/v${version}"
mv "releases/v${version}/bin/helmpull" "bin/helmpull" || \
    mv "releases/v${version}/bin/helmpull.exe" "bin/helmpull"