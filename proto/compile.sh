#!/bin/bash

cd attributes
./compile.sh
cd ..

cd auth
./compile.sh
cd ..

cd authority
./compile.sh
cd ..

cd config
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


cd users
./compile.sh
cd ..