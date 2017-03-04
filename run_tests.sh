#!/usr/bin/env bash
retStatus=0
for module_dir in */; do
    module_dir=${module_dir%*/}
    echo ""
    echo "############# TESTING $module_dir ###############"
    echo ">>>> Generating pool"
    yep generate -t ./$module_dir 2>/dev/null
    let "retStatus=retStatus + $?"
    echo ""
    echo ">>>> Executing tests"
    go test -v ./$module_dir/...
    let "retStatus=retStatus + $?"
done
exit $retStatus
