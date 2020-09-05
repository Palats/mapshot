#!/bin/bash
echo "-- Automatically generated, do not modify" > generated.lua
echo "local data = {}" >> generated.lua
echo "data.html = [==[" >> generated.lua
cat viewer.html >> generated.lua
echo "]==]" >> generated.lua
echo "return data" >> generated.lua