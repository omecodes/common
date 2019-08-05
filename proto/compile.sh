#!/bin/bash

cd attributes
./compile.sh
cd ..

cd authority
./compile.sh
cd ..

cd registry
./compile.sh
cd ..

cd supervisor
./compile.sh
cd ..

cd tree
./compile.sh
cd ..