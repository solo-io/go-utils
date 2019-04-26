# Manifest Test Library

This package contains some utils for easily parsing installation manifests (i.e. Helm-generated yaml files) 
and writing tests against those manifests. This is to ensure your Helm chart is properly linted and meets 
expectations. 

## Usage

- Define a test suite that parses a manifest, for example see `test/example_suite_test.go`. 
- Define a test file that tests the manifest, for example see `test/example_test.go`. 

These tests should be very concise based on the resource builder utilities provided.

## Future

In the short term, the goal is increasing test coverage of helm charts that are published by different 
projects. Some useful extensions in the future would include:

- Extra utilities for helping test helm charts with different values overrides. 

It may make sense to turn this into a Helm chart generator in the future, to the extent that process can be
simplified (see `github.com/solo-io/build` for more info about the Solo common build SDK). 