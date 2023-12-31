/*
Copyright (c) 2018-2019 The sedracoin developers
Copyright (c) 2013-2018 The btcsuite developers
Copyright (c) 2015-2016 The Decred developers
Copyright (c) 2013-2014 Conformal Systems LLC.
Use of this source code is governed by an ISC
license that can be found in the LICENSE file.

sedrad is a full-node sedra implementation written in Go.

The default options are sane for most users. This means sedrad will work 'out of
the box' for most users. However, there are also a wide variety of flags that
can be used to control it.

Usage:

	sedrad [OPTIONS]

For an up-to-date help message:

	sedrad --help

The long form of all option flags (except -C) can be specified in a configuration
file that is automatically parsed when sedrad starts up. By default, the
configuration file is located at ~/.sedracoin/sedrad.conf on POSIX-style operating
systems and %LOCALAPPDATA%\sedracoin\sedrad.conf on Windows. The -C (--configfile)
flag can be used to override this location.
*/
package main
