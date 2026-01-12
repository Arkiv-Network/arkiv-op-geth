#!/usr/bin/env bash

MONGO_URI="mongodb://admin:password@localhost:27017/"

exec mongosh "$MONGO_URI"
