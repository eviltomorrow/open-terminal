#!/bin/bash

# set -eo pipefail

os=$(uname -s)
GOOS=""
case ${os} in
"Linux")
    GOOS="linux"
    ;;
"Darwin")
    GOOS="darwin"
    ;;
*)
    echo -e "[\033[34mFatal\033[0m]: 暂不支持的系统类型[${os}] "
    exit 255
    ;;
esac

arch=$(uname -m)
GOARCH=""
case ${arch} in
"x86_64")
    GOARCH="amd64"
    ;;
*)
    echo -e "[\033[34mFatal\033[0m]: 暂不支持的 cpu 架构[${arch}] "
    exit 255
    ;;
esac

root_dir=$(pwd)
PATH=${root_dir}/tools/protoc-gen-go/${GOOS}_${GOARCH}:$PATH

for name in $(ls ${root_dir}/lib/entity); do
    # echo ${name}
    if [ "${name##*.}"x = "proto"x ]; then
        mkdir -p ${root_dir}/lib/grpc/pb/entity
        
        file_name=${root_dir}/lib/entity/${name}

        ${root_dir}/tools/protoc/${GOOS}_${GOARCH}/bin/protoc --proto_path="${root_dir}" --proto_path="${root_dir}/lib/entity" --go_opt=paths=source_relative --go_out="${root_dir}/lib/grpc/pb/entity" --go-grpc_opt=paths=source_relative --go-grpc_out="${root_dir}/lib/grpc/pb/entity" ${name}
        code=$(echo $?)
        if [ $code = 0 ]; then
            echo -e "编译文件: ${file_name} => [\033[32m成功\033[0m] "
        else
            echo -e "[\033[34mFatal\033[0m]: 编译文件: [${file_name}] => [\033[34m失败\033[0m] "
            echo -e "\t <<<<<<<<<<<< 编译过程意外退出，已终止  <<<<<<<<<<<<"
            exit
        fi
    fi
done



for name in $(ls ${root_dir}/apps); do
    # echo ${name}
    if echo "${name}" | grep -q -E '\-web$'
    then
        continue
    fi

    app_dir=${root_dir}/apps/${name}
    app_name=${name}
    adapter_dir=${app_dir}/adapter
    if [ -d ${adapter_dir} ]; then
        ok=false
        for name in $(ls ${adapter_dir}); do
            if [ "${name##*.}"x = "proto"x ]; then
                ok=true
                break
            fi
        done

        if [ ${ok}=true ]; then
            pb_dir=${root_dir}/lib/grpc/pb/${app_name}
            if [ -d ${pb_dir} ]; then
                rm -rf ${pb_dir}
            fi
            mkdir -p ${pb_dir}

            for name in $(ls ${adapter_dir}); do
                file_name=${adapter_dir}/${name}

                ${root_dir}/tools/protoc/${GOOS}_${GOARCH}/bin/protoc --proto_path="${root_dir}" --proto_path="${adapter_dir}" --go_opt=paths=source_relative --go_out="${pb_dir}" --go-grpc_opt=paths=source_relative --go-grpc_out="${pb_dir}" ${name}
                code=$(echo $?)
                if [ $code = 0 ]; then
                    echo -e "编译文件: ${file_name} => [\033[32m成功\033[0m] "
                else
                    echo -e "[\033[34mFatal\033[0m]: 编译文件: [${file_name}] => [\033[34m失败\033[0m] "
                    echo -e "\t <<<<<<<<<<<< 编译过程意外退出，已终止  <<<<<<<<<<<<"
                    exit
                fi
            done
        fi
    fi
done

