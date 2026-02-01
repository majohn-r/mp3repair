# mp3repair

[![GoDoc Reference](https://godoc.org/github.com/majohn-r/mp3repair?status.svg)](https://pkg.go.dev/github.com/majohn-r/mp3repair)
[![go.mod](https://img.shields.io/github/go-mod/go-version/majohn-r/mp3repair)](go.mod)
[![LICENSE](https://img.shields.io/github/license/majohn-r/mp3repair)](LICENSE)

[![Release](https://img.shields.io/github/v/release/majohn-r/mp3repair?include_prereleases)](https://github.com/majohn-r/mp3repair/releases)
[![Code Coverage Report](https://codecov.io/github/majohn-r/mp3repair/branch/main/graph/badge.svg)](https://codecov.io/github/majohn-r/mp3repair)
[![Go Report Card](https://goreportcard.com/badge/github.com/majohn-r/mp3repair)](https://goreportcard.com/report/github.com/majohn-r/mp3repair)
[![Build Status](https://img.shields.io/github/actions/workflow/status/majohn-r/mp3repair/build.yml?branch=main)](https://github.com/majohn-r/mp3repair/actions?query=workflow%3Abuild+branch%3Amain)

## What is **mp3repair**?
**mp3repair** is an application that finds and repairs mismatches between mp3 file metadata and the files' names and
containing folders.
## Why do I need **mp3repair**?
Sometimes you rip a CD and expect the files to play in the same order they appear on the CD, and they don't. I wrote
the **mp3repair** program to look for common causes of this problem and to repair them.
### How does it work?
Quite well, thank you for asking.

OK, seriously, let's start with what happens when you rip a CD. The ripping application gets the name of each track,
the name of the album, and the name of the recording artist (pop music, for example) or composer (classical music, for
example). It ensures that a folder exists that corresponds to the name of the recording artist or composer. It ensures
that the artist folder contains a subfolder corresponding to the album name. It ensures that the album subfolder
contains mp3 files whose names consist of the track number and the track name. _It also embeds all that naming
information into the mp3 file as a block of metadata._ That's _important_ - we'll come back to that.

Now you might wonder: how did the ripping application know those names? They're not on the CDs themselves (ok,
sometimes, they are, but too rarely to be a dependable source). When the application scans the CD, it discovers how
many tracks there are and how long each one is. It then checks an internet service (there are several of them) and asks
that service, "I have a CD with this many tracks, and here are the track lengths. Tell me about that CD." Usually,
that's enough information for the service to uniquely identify the CD, and the service replies with what it knows,
including, at a minimum, the track, album, and artist names.

How did the service know those names? People all over the world provide that information. _**Sometimes that information
is wrong!**_

The names are _**crowdsourced**_. How could it **not** go wrong?

That's why your car's audio system, for example, can't make sense of some of the albums you took on that road trip.
### If the file and folder names are wrong, I can fix that with File Explorer!
Yes, you can, and **you need to**. If you don't rename and rearrange the incorrectly written files and folders, there
is not much **mp3repair** can do - it **_assumes_** that the file and folder names are correct. So you rename files and
folders and drag files that have been scattered across multiple folders into a single correctly named folder, and the
application you use to play the files will probably still be confused. It'd be a pleasant world if that application
would simply honor how the files are named and organized. Unfortunately, they typically don't.

Remember that the ripping application embedded those names in a block of metadata inside each mp3 file? The playing
application assumes that metadata is the truth, and that's how it knows what to play and in what order when you tell it
"play such-and-such album".

Rearranging the files does not affect their metadata.

That's what **mp3repair** is for.

**mp3repair** scans the music folder, collecting artist, album, and track names, reading the metadata from each mp3
file, and seeing where they don't match. **mp3repair** can then _**rewrite**_ each mp3 file's metadata to match the
file and folder names. That's why you need to use File Explorer or an equivalent application to change the names or
folder contents. If you don't, they'll match the mp3 files' metadata - the ripping application wrote them both, after
all - and **mp3repair** will not see a mismatch.

After **mp3repair** has rewritten their metadata, the files should play in the order you expect them to.
