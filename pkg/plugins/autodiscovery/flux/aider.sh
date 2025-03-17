#!/bin/bash

export OLLAMA_API_BASE=http://127.0.0.1:11434

export

#aider --model ollama_chat/deepseek-r1:1.5b
#aider --model ollama_chat/deepseek-coder:6.7b


source .env
aider --model deepseek --api-key deepseek=$DEEPSEEK_API_KEY
