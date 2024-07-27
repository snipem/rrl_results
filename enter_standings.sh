#!/bin/bash

url="$1"
result_file_name="results.csv"
curl -s "$1" > input.html

echo "$1" > $result_file_name

while true 
do
    cat input.html | 
        pup ".jsParticipant strong text{}" | 
	sed -e 's/^[[:space:]]*//g' |
        fzf --bind=enter:replace-query+print-query >> $result_file_name || break
done

