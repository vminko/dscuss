#!/bin/sh
#
# This file is used for generating documentation via godoc tool.
# Usage is described in the source code below.
#
# Copyright (c) 2018 Vitaly Minko <vitaly.minko@gmail.com>

EXCLUDED_DIRS="docs utils"
STATIC_DIR="$GOPATH/src/golang.org/x/tools/godoc/static/"

die() {
    echo "Error: $1"
    exit 1
}

generate_htmls() {
    local package=$1
    local target_dir=$2
    local target_uri=$3

    local excludes=""
    for exclude in ${EXCLUDED_DIRS}; do
        excludes="${excludes} -not -path '${exclude}' -not -path '${exclude}/*'"
    done

    mkdir -p ${target_dir} || die "failed to create directory ${target_dir}"
    echo "Generating documentation for ${package}..."
    godoc -url "/pkg/${package}/" > "${target_dir}/index.html" || \
        die "failed to generate doc for ${package}"

    local subpackages=$(cd ${GOPATH}/src/${package} && echo ${excludes} | xargs find * -type d)
    for subpackage in  ${subpackages}; do
        local target="${target_dir}/${subpackage}"
        mkdir -p ${target} || die "failed to create directory ${target}"
        echo "Generating documentation for ${package}/${subpackage}..."
        godoc -url "/pkg/${package}/${subpackage}" > "${target}/index.html" || \
            die "failed to generate doc for ${package}/${subpackage}"
    done

    # Replace target URI
    find ${target_dir} -name index.html -exec \
        sed -i "s|/lib/godoc/|${target_uri}${target_dir}/|" {} \; || \
        die "failed to replace target URI in a static"
}

copy_statics() {
    local target_dir=$1
    local target_uri=$2
    local statics=`grep "${target_uri}${target_dir}" ${target_dir}/index.html | \
                   sed -n "s|.*${target_uri}${target_dir}\(.*\)\".*|\1|p"`
    for static in  ${statics}; do
        cp ${STATIC_DIR}${static} ${target_dir} || die "failed to copy a static"
    done
}

if [ $# -ne 3 ]
then
    cat >&2 << EOF
Usage:
    $0 <package> <target_dir> <host_prefix>
where,
    <package> is the root package to generate documentation for.
    <target_dir> is the directory to put the result in.
    <target_uri> is the URI where the result will be published.
Example:
    utils/generate_docs.sh vminko.org/dscuss godoc /storage/dscuss/
EOF
    exit 1
fi

PACKAGE=$1
TARGET_DIR=$2
TARGET_URI=$3

if [ -z "${GOPATH}" ]; then
	echo "\$GOPATH is not set." 1>&2
fi

if [ ! -d "${STATIC_DIR}" ]; then
	echo "Static directory of the godoc tool not found: ${STATIC_DIR}." 1>&2
fi

generate_htmls ${PACKAGE} ${TARGET_DIR} ${TARGET_URI}
copy_statics ${TARGET_DIR} ${TARGET_URI}
