#!/bin/sh

# Generate 50 keys and have them signed by the CA and time how long this takes
# for RSA, ECDSA and EdDSA.
#
# On my oldish laptop:
#
# $ sh benchmark.sh 
# RSA 53s
# ECDSA 1s
# EdDSA 0s

for TYPE in RSA ECDSA EdDSA; do
	mkdir "${TYPE}"
	CA_DIR="${TYPE}" CA_KEY_TYPE="${TYPE}" _bin/vpn-ca -init-ca -name "${TYPE} CA"
	START=$(date +%s)
	for i in $(seq 50); do
		CA_DIR="${TYPE}" CA_KEY_TYPE="${TYPE}" _bin/vpn-ca -client -name "${TYPE}-client-${i}"
	done
	END=$(date +%s)
	echo "${TYPE} $((END-START))s"
	rm -rf "${TYPE}"
done
