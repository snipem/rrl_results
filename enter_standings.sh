
curl -s "$1" > input.html

while true 
do
    cat input.html | 
        pup ".jsParticipant strong text{}" | 
        fzf --bind=enter:replace-query+print-query >> results.txt || break 
done

