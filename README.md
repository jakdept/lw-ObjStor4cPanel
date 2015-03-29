cPanelToObjectStore
===================

**_This is experimental for the time being_**

This is a custom transport script that will allow cPanel/WHM backups to be sent to Liquid Web Object Storage.

I assume the user understands the concept of access keys, secret keys, buckets, and has a bucket already set up in Object Storage.

## Requirements ##

This is written in Go. In order to compile this, you must have Go installed.

Eventually, a compiled binary will be available to copy to the host system, and simply run. We may also choose to go with a RPM package instead.

## Usage ##

1. Compile the binary.
1. Copy the compiled binary to somewhere on the server, and make sure it's executable.
1. Go to *Additional Settings* in the Backup Configuration in WHM and create a *Custom* destination
1. Give it a name and select whether you want to transfer system files
1. For `Script`, point it to the place you stuck the script
1. `Backup Directory` is optional, but good to use if the bucket you're going to use already has other things
1. `Remote Host` is the **name of the bucket** you are going to use
1. `Remote Account Username` is going to be a valid access key that you have setup
1. `Remote Password` is going to be the secret key that is associated with the access key that you're using

