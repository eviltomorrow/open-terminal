#!/bin/bash

set -eo pipefail 

root_dir=$(pwd)
app_dir=${root_dir}/apps
bin_dir=${root_dir}/bin

MAINVERSION=$(cat ${root_dir}/version)
GITSHA=$(git rev-parse HEAD)
BUILDTIME=`date +%FT%T%z`
gopaths=(${GOPATH//:/ })
TRIMGOPATH=""
let length=${#gopaths[@]}-1
for((i=0;i<${#gopaths[@]};i++)) 
do
    if [ ${i} = ${length} ]; then
        TRIMGOPATH="${TRIMGOPATH} -trimpath=${gopaths[i]}/src"
    else
        TRIMGOPATH="-trimpath=${gopaths[i]}/src ${TRIMGOPATH}"
    fi
done

GCFLAGS="all=${TRIMGOPATH}"

CGO_ENABLED=$(echo $CGO_ENABLED)
RACEFLAGS=""
if [ "${CGO_ENABLED}" == "1" ]; then
    RACEFLAGS="-race"
fi

function build_app(){
    LDFLAGS="-X main.AppName=${1} -X main.MainVersion=${MAINVERSION} -X main.GitSha=${GITSHA} -X main.BuildTime=${BUILDTIME} -s -w"

    echo -e "\033[32m=> Building binary(${1})...\033[0m"
    mkdir -p ${bin_dir}/${1}

    if [ -d ${app_dir}/${1}/conf/etc ]; then
        cp -rp ${app_dir}/${1}/conf/etc ${bin_dir}/${1}
    fi

    echo "go build -ldflags "${LDFLAGS}" -gcflags "${GCFLAGS}" -o ${bin_dir}/${1}/bin/${1} ${app_dir}/${1}/main.go"
    go build ${RACEFLAGS} -ldflags "${LDFLAGS}" -gcflags "${GCFLAGS}" -o ${bin_dir}/${1}/bin/${1} ${app_dir}/${1}/main.go

    echo -e "\033[32m=> Build Success\033[0m"
}

if [ ! -z "${1}" ]; then
    build_app ${1}
else
    for name in $(ls ${app_dir}); do
        if [ -e ${app_dir}/${name}/.build_skip ]; then
            continue
        fi
        build_app ${name}
    done
fi
