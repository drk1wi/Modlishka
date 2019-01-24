#!/bin/sh
set -e

RUN_CMD='${MODLISHKA_BIN}'
IFS='
'
envList=$(env)
for line in $envList
do
        envName=$(echo "${line}" | cut -d'=' -f 1 | tr '[:upper:]' '[:lower:]')
        if [[ "${envName:0:3}" = "ml_" ]] ;
        then
			optionName=$(echo "${envName}" | sed -r 's/(^ml\_)//g' |awk -F'_' '{ printf $1; for(i=2; i<=NF; i++) printf toupper(substr($i,1,1)) substr($i,2);printf "\n"}')
			optionValue=$(echo "${line}" | sed -r 's/(^.*)\=//g')
			RUN_CMD="${RUN_CMD} -${optionName} ${optionValue}"
        fi
done

echo "Running command: ${RUN_CMD}"

sh -c $RUN_CMD