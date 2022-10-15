# exfs

TODO: rename to exfs - extra filesystem commands

The motivation of this repository is to test recurring filesystem operations that WOULD NOT appear in the standard library.

Between `os`, `path` & `ioutil` go doesn't leave much to be desired in terms of interacting with the underlying operating system.

However, as humans, we ofter encounter more specific, extensive patterns that require testing.

This project should not have dependencies and the function names should be complete and human readable.

# test

> make test
