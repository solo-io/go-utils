# go-utils

This repo contains common utilities for go projects.


## Structure

go-utils currently contains 2 main types of packages: libraries, and utilities. Utility package names
contain the suffix `utils` while library packages do not. 

The difference between these 2 is subtle and definitely up for debate. A library is a package which is
meant to serve a single purpose, or expose a specific functionality. While utilities are groups of
semi-related functions which may accomplish many different ends.


`ci`: files and objects used for testing in CI.
