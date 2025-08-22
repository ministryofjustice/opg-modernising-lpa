for k in `cat lang/en.json | jq -r 'keys|join("\n")'`;
do
    # Get the part of the key before the first colon, if it exists, otherwise use the key as is
    search_pattern="${k%%:*}"

    COUNT=$(git grep "$search_pattern" | grep -vE '(\.json:|\_test.go:|cy.js:)' | wc -l)

    if [ "$COUNT" -eq 0 ]; then
        echo "$k"
    fi
done
