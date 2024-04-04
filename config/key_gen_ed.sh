#!/bin/bash

RED='\033[0;31m'
NC='\033[0m'

if [[ $# -lt 1 ]]
then
    echo -e "${RED}Usage: $0 <number of nodes>${NC}"
    echo "Example: $0 100"
    exit 1
fi

NUMNODES=$1
KEYPATH="keys"

# Remove existing keys.
rm -rf $KEYPATH

# Create directory for new keys.
mkdir -p $KEYPATH
exitcode=$?
if [[ $exitcode -ne 0 ]] || [[ ! -d $KEYPATH ]]
then
    echo -e "${RED}Key directory $KEYPATH cannot be accessed!${NC}"
    exit 1
fi

for i in $(seq 1 $NUMNODES)
do
    j=$((i-1))
    PRIVKEYFILE="$KEYPATH/$j.priv"
    PUBKEYFILE="$KEYPATH/$j.pub"

    (
     # Generating Ed25519 key pairs
     openssl genpkey -algorithm Ed25519 -out $PRIVKEYFILE
     openssl pkey -in $PRIVKEYFILE -pubout -out $PUBKEYFILE
    ) &
done

wait

printf "${RED}$NUMNODES keys created!${NC}\n"
