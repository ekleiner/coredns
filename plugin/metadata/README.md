# metadata

## Name

*metadata* - enables a metadata.

## Description

By enabling *metadata* any plugin that implements [metadata.metadater interface](https://godoc.org/github.com/coredns/coredns/plugin/metadata#Metadater) will be called and will be able to add it's own Metadata to context. Metadata is available for all plugins via Context.

## Syntax

~~~
metadata
~~~

## Plugins

Any plugin that implements the Metadater interface will be called to get metadata variables.

## Examples

Enable metadata

~~~ corefile
. {
    metadata
}
~~~
