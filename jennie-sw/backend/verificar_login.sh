#!/bin/bash
USUARIO=$1
SENHA_DIGITADA=$2

LINHA_SHADOW=$(grep "^${USUARIO}:" /etc/shadow)

if [ -z "$LINHA_SHADOW" ]; then
    echo "erro"
    exit 1
fi

HASH_SALVO=$(echo "$LINHA_SHADOW" | cut -d: -f2)

if [[ "$HASH_SALVO" == "*" ]] || [[ "$HASH_SALVO" == "!" ]]; then
    echo "erro"
    exit 1
fi

RESULTADO=$(perl -e 'print (crypt($ARGV[0], $ARGV[1]) eq $ARGV[1] ? "sucesso" : "erro")' "$SENHA_DIGITADA" "$HASH_SALVO")

echo "$RESULTADO"