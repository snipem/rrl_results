#!/bin/bash

echo URL:
read url

echo Platz des ersten Fahreres mit DNF. 0 wenn es kein DNF gab:
read dnf

result_file_name="results_$(basename $url).csv"

curl -s "$url" > input.html

echo "$url" > $result_file_name
echo "$dnf" >> $result_file_name

while true 
do
    cat input.html | 
        pup ".jsParticipant strong text{}" | 
	sed -e 's/^[[:space:]]*//g' |
        fzf --bind=enter:replace-query+print-query >> $result_file_name || break
done

go run format.go --results "$result_file_name"

