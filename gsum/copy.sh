#!/bin/bash

names=$(cat <<EOF
forward1.input.txt
forward1.output.txt
forward2.input.txt
forward2.output.txt
forward3.input.txt
forward3.output.txt
forward4.input.txt
forward4.output.txt
from.input.txt
from.output.txt
from2.input.txt
from2.output.txt
from3.input.txt
from3.output.txt
negative.input.txt
negative.output.txt
someoneWrote1.input.txt
someoneWrote1.output.txt
someoneWrote2.input.txt
someoneWrote2.output.txt
EOF
)
for n in $names
do 
    curl -o testdata/$n https://raw.githubusercontent.com/Like-Falling-Leaves/parse-reply/master/test/files/$n
done

