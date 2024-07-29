#!/bin/bash

cd $HOME/work/rrl_results

url=$(curl --silent "https://rookie-racing.eu/kalender/calendar-feed/" | grep -Eo "(http|https)://[a-zA-Z0-9./?=_%:-]*" | sort -u | grep "/event/" | fzf --prompt "URL: > ")

echo Platz des ersten Fahreres mit DNF. 0 wenn es kein DNF gab:
read dnf

result_file_name="results/results_$(basename $url).csv"

curl -s "$url" > input.html

echo "$url" > $result_file_name
echo "$dnf" >> $result_file_name

fzf_prompt="Schnellste Runde > "
i=1

while true 
do
    cat input.html | 
        pup ".jsParticipant strong text{}" | 
        sed -e 's/^[[:space:]]*//g' |
        sed -e 's/[[:space:]]*$//g' |
        cat - <(cat einteilung.csv | cut --delimiter=, --fields 2) |
        sort -u |
        fzf --prompt="$fzf_prompt" -i --bind=enter:replace-query+print-query >> $result_file_name || break
        fzf_prompt="CTRL-C -> quit: P$i > "
        i=$(( $i + 1 ))
done

echo ""

go run format.go --results "$result_file_name" | tee whatsapp.txt

