# Release Process

Boilersuite is an internal tool for the cert-manager org / project, and as such is ad-hoc and largely unsupported for any other use.

For this reason releases are very basic and the release process is not automated.

1. Create a git tag with the desired version number (`git tag v0.x.y`)
2. Push the tag to the main repository (not a fork) (`git push origin v0.x.y`)
3. Locally, run `make build-release`
4. Create a release on GitHub using the tag, and add auto generated GitHub release notes.
5. Upload the build artifacts on the GitHub release page
    1. boilersuite-darwin-amd64
    2. boilersuite-darwin-arm64
    3. boilersuite-linux-amd64
    4. SHA256SUMS
6. Publish the release
