#!/bin/bash

# Coleta temperatura
temperatura_arquivo=$(cat /sys/class/thermal/thermal_zone0/temp 2>/dev/null)
if [ -z "$temperatura_arquivo" ]; then
    temperatura="45.0"
else
    temperatura=$(awk -v t="$temperatura_arquivo" 'BEGIN {printf "%.1f", t/1000}')
fi

# Coleta uso de memória RAM em mb
ram_usada=$(free -m | awk '/Mem:/ {print $3}')
ram_total=$(free -m | awk '/Mem:/ {print $2}')

# Coleta o quanto de armazenamento está sendo usado na partição raiz
disco_porcentagem=$(df -h / | awk 'NR==2 {print $5}')

echo "$temperatura,$ram_usada,$ram_total,$disco_porcentagem"