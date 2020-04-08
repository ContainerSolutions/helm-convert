#!/bin/sh

set -e

current_version=$(sed -n -e 's/version:[ "]*\([^"]*\).*/\1/p' $(dirname $0)/plugin.yaml)
HELM_CONVERT_VERSION=${HELM_CONVERT_VERSION:-$current_version}

dir=${HELM_PLUGIN_DIR:-"$(helm home)/plugins/helm-convert"}
os=$(uname -s | tr '[:upper:]' '[:lower:]')
release_file="helm-convert_${os}_${HELM_CONVERT_VERSION}.tar.gz"
url="https://github.com/deedubs/helm-convert/releases/download/v${HELM_CONVERT_VERSION}/${release_file}"

mkdir -p $dir

if command -v wget
then
  wget -O ${dir}/${release_file} $url
elif command -v curl; then
  curl -L -o ${dir}/${release_file} $url
fi

tar xvf ${dir}/${release_file} -C $dir

chmod +x ${dir}/helm-convert

rm ${dir}/${release_file}
