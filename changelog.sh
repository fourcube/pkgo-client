#!/usr/bin/env bash
git log --oneline $(git describe --tags --abbrev=0 @^)..@ --pretty=format:"- %s (%h)" | pbcopy
