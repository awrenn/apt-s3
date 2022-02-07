# apt-s3

`apt-s3` is a little command line tool for adding .deb files to an S3 bucket, so that they can be consumed via apt without having to run a webserver.
It is intended to "compete" with https://github.com/deb-s3/deb-s3/, which uses ruby - I was originally a bit annoyed with having to get the right version of ruby installed, which I don't do often. A single static go binary will be easier for me to install myself. Obviously if that doesn't bother you, contribute and keep using deb-s3, but why not scratch your own itch over a weekend?

It currently only has very basic features - if you point to a deb file, it will:

1. Extract the control file from this deb package
2. Upload the .deb file to the asset pool, without indexing it.
3. Download the existing package file for repo/codename/arch, add this deb file and its hashes, and re-upload the package file.
4. Download the existing Release file, adds the package's arch and component if not present, and then re-uploads the Release file.

`apt-s3` is still missing a couple major features, the biggest one being:

1. The ability to change the component that the package is uploaded under - currently only `main` for a component and `stable` for a codename.

The following command can be used to a deb file to a bucket:  
`./apt-s3 -region us-west-2 -bucket dummy-apt -deb ./dummy.deb`  


Some other features I am hoping to get added, besides the two obvious missing features above:   
1. The ability to correct/fix existing Package and Release files.  
2. The CLI experience is quite poor, due to using `flag` instead of a 3rd party library. Smarter usage of the flag library / simple first-party cli tooling is pretty curcial.  
3. Eliminate 3rd party dependencies - currently there are three direct libraries - a go XZ library, AWS's s3 library, and a reliance on the system's `ar` binary.   
4. Better automated testing - currently, tests are best ran with `go test -v`, and manually inspecting outputs.  
