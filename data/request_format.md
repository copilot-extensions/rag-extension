Requests to your agent will be `application/json` formatted.  An example curl request to your agent may look like this:

```
curl --request POST \
    --url $AGENT_URL \
    --header 'Accept: application/json' \
    --header 'Content-Type: application/json' \
    --header "X-GitHub-Token: $RUNTIME_GENERATED_TOKEN" \
    --data '{
        "messages": [
            {
                "role": "user",
                "content": "What is a closure in javascript?",
                "copilot_references": []
            }
        ]
    }'
```
