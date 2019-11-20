#!/bin/sh

TAG=$(git describe --abbrev=0 --tags)
OUTDIR=output/vpn-ca-${TAG}
mkdir -p ${OUTDIR}

make clean
GOOS=darwin GOARCH=amd64 make
mkdir -p ${OUTDIR}/mac
cp _bin/vpn-ca ${OUTDIR}/mac

make clean
GOOS=windows GOARCH=amd64 make
mkdir -p ${OUTDIR}/windows
cp _bin/vpn-ca ${OUTDIR}/windows/vpn-ca.exe

make clean
GOOS=linux GOARCH=amd64 make
mkdir -p ${OUTDIR}/linux
cp _bin/vpn-ca ${OUTDIR}/linux

(
    cd output
    zip -r ../vpn-ca-${TAG}.zip vpn-ca-${TAG}
)

minisign -Sm vpn-ca-${TAG}.zip
