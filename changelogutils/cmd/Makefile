
README.md:
	echo "This is a sample of the output from the changelog generation script.\n" > $@
	echo "To update this file, run: \`cd changelogutils/cmd && make README -B && cd -\`\n" >> $@
	echo "---" >> $@
	cd ../.. && go run changelogutils/cmd/main.go >> changelogutils/cmd/README.md
