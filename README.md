# go-utils

This repo contains common utilities for go projects.


## Structure

`lib`: packages which should serve as mini libraries in and of themselves. The reason I seperated these from utils is that the utils folders are collections of many semi-related, but ultimately different helper functions. While, the common packages each perform a specific function, and expose it as a library. These packages names are simple, and explain the use.

`pkg`: packages which contain groups of utility functions for specific purposes. These packages names each end with utils to clearly distinguish them as such.

`test`: packages which contain testing information for this repo. Not created to be use by outside projects

`ci`: files and objects used for testing in CI.