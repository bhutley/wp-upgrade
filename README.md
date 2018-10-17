# wp-upgrade

## Upgrade Wordpress from the latest source code

To do an upgrade, download the latest version of Wordpress and extract it to a directory.
Then run wp-upgrade as

`wp-upgrade -wp-dir path-to-latest-source -site-dir path-to-existing-site`

This will create any new directories in the latest source code, and copy all the files over.

TODO: Remove any left-over files that have since been removed.
 