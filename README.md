# webbscan

This is intended to be a simple(ish) port scanner tool, written as a
demonstration of my programming abilities and as an exercise in learning the
Go programming language.

The source code is divided up in the several packages, nominally to maximize
reusability of each module:

portserv - a simple library for translating TCP and UDP port numbers into
	   service names, using /etc/services
progbar	 - a simple ASCII-text-based "progress bar" which gives the user
	   feedback during the portscan, visually displaying the progress of
	   the scan, as well as a "spining" effect showing that probes are
	   actively being sent.
tcpProbe - a simple library for probing sockets.
vdiag	 - a handy package for producing diagnositic output.
workflow - a library for administering task execution, supporting concurrent
	   as well as rate-limited dispatching.
webbscan - the port scanner tool.


Highlights of the code include:
  - understanding of various Go syntax, constructions, and usage, including
    arrays, maps, slices, ranges, structures, methods, strings,
    initialization, interfaces, enumerations (using iota), bytes-Buffers,
    Writers, command line flag support, goroutines, channels, timers, 
    functions-as-parameters, and closures
  - regular-expression-based parsing
  - concurrency and synchronization
  - Go unit testing support, test-case objects, use of interfaces for mocking
    functions


The tool provides several command-line switches which control its execution:

    `-agents` (default 8):  the number of concurrent probes
    `-host` (default "127.0.0.1"):  the target host to probe
    `-protocol` (default "tcp"):  Protocol ("tcp" or "udp")
    `-rate` (default unlimited):  the maximum number of probes to be sent per
				  second (0: unlimited)
    `-verbose` (default none):	The level of verbosity for diagnostic messages
				(`-v` is a shorthand for "level 2")

In addition to the tool source code, the source includes unit tests for
(nearly) all functions.
